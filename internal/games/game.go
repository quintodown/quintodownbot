package games

import (
	"errors"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/mailru/easyjson"
	"github.com/quintodown/quintodownbot/internal/app"
	"github.com/quintodown/quintodownbot/internal/pubsub"
)

type GameInfoClient interface {
	GetGames(Competition) ([]Game, error)
	GetGameInformation(Competition, string) (Game, error)
}

type GameHandler struct {
	client   GameInfoClient
	gameList sync.Map
	queue    pubsub.Queue
	clk      app.Clock
}

func NewGameHandler(client GameInfoClient, getGames bool, queue pubsub.Queue, clk app.Clock) *GameHandler {
	gh := &GameHandler{client: client, queue: queue, clk: clk}

	if getGames {
		go func() { gh.initGames() }()
	}

	return gh
}

func (gh *GameHandler) GetGames(c Competition) []Game {
	found := gh.loadGames(c)

	sort.Slice(found, func(i, j int) bool {
		return found[i].Start.Before(found[j].Start)
	})

	return found
}

func (gh *GameHandler) GetGamesStartingIn(c Competition, d time.Duration) []Game {
	var found []Game

	offset := gh.clk.Now().Add(d).Truncate(time.Minute)

	for _, v := range gh.loadGames(c) {
		if offset.Equal(v.Start.UTC().Truncate(time.Minute)) {
			found = append(found, v)
		}
	}

	return found
}

func (gh *GameHandler) GetGame(id string) (Game, error) {
	for _, competition := range GetCompetitions() {
		for _, v := range gh.loadGames(competition) {
			if v.Id == id {
				return v, nil
			}
		}
	}

	return Game{}, errors.New("game not found")
}

func (gh *GameHandler) UpdateGamesInformation(onlyPlaying bool) {
	for _, competition := range GetCompetitions() {
		gameList := gh.loadGames(competition)
		for i, v := range gameList {
			if onlyPlaying && !v.isGameInProgress(gh.clk) {
				continue
			}

			g, err := gh.client.GetGameInformation(v.Competition, v.Id)
			if err != nil {
				continue
			}

			lastGameChange := gh.getLastGameChange(v, g)
			switch lastGameChange {
			case Rescheduled:
				gameList[i].Start = g.Start
				gameList[i].Status.State = RescheduledState
			case HomeScore:
				gameList[i].HomeTeam.Score = g.HomeTeam.Score
			case AwayScore:
				gameList[i].AwayTeam.Score = g.AwayTeam.Score
			case Started:
				gameList[i].Status.Period = 1
				gameList[i].Status.DisplayClock = g.Status.DisplayClock
			case PeriodFinished:
				gameList[i].Status.Period = g.Status.Period
				gameList[i].Status.DisplayClock = g.Status.DisplayClock
			case GameFinished:
				gameList[i].Status = g.Status
			}

			if lastGameChange != NoChanges {
				_ = gh.queue.Publish(
					pubsub.GamesTopic.String(),
					message.NewMessage(
						watermill.NewUUID(),
						gameList[i].toGameEvent(lastGameChange),
					),
				)
			}
		}

		gh.gameList.Store(competition.String(), gameList)
	}
}

func (gh *GameHandler) initGames() {
	for _, v := range GetCompetitions() {
		games, err := gh.client.GetGames(v)
		if err != nil {
			continue
		}

		gh.gameList.Store(v.String(), games)
	}
}

func (gh *GameHandler) loadGames(c Competition) []Game {
	f, ok := gh.gameList.Load(c.String())
	if !ok {
		return nil
	}

	return f.([]Game)
}

func (gh *GameHandler) getLastGameChange(oldGameInfo, newGameInfo Game) GameChange {
	lastGameChange := NoChanges

	if reflect.DeepEqual(oldGameInfo, newGameInfo) {
		return lastGameChange
	}

	if !newGameInfo.Start.Equal(oldGameInfo.Start) && oldGameInfo.Status.State != RescheduledState {
		lastGameChange = Rescheduled
	}

	if newGameInfo.HomeTeam.Score != oldGameInfo.HomeTeam.Score {
		lastGameChange = HomeScore
	}

	if newGameInfo.AwayTeam.Score != oldGameInfo.AwayTeam.Score {
		lastGameChange = AwayScore
	}

	if newGameInfo.Status.Period != oldGameInfo.Status.Period && newGameInfo.Status.State == InProgressState {
		if newGameInfo.Status.Period == 1 {
			lastGameChange = Started
		} else {
			lastGameChange = PeriodFinished
		}
	}

	if newGameInfo.Status.State == FinishedState && oldGameInfo.Status.State != FinishedState {
		lastGameChange = GameFinished
	}

	return lastGameChange
}

func (g *Game) toGameEvent(lastGameChange GameChange) []byte {
	mb, _ := easyjson.Marshal(pubsub.GameEvent{
		Id:    g.Id,
		Start: g.Start,
		Name:  g.Name,
		Venue: struct {
			FullName string `json:"full_name"`
			City     string `json:"city"`
			State    string `json:"state"`
			Capacity int    `json:"capacity"`
			Indoor   bool   `json:"indoor"`
		}{
			FullName: g.Venue.FullName,
			City:     g.Venue.Address.City,
			State:    g.Venue.Address.State,
			Capacity: g.Venue.Capacity,
			Indoor:   g.Venue.Indoor,
		},
		Status: struct {
			Clock        float64 `json:"clock"`
			DisplayClock string  `json:"display_clock"`
			Period       int     `json:"period"`
			State        string  `json:"state"`
		}{
			Clock:        g.Status.Clock,
			DisplayClock: g.Status.DisplayClock,
			Period:       g.Status.Period,
			State:        g.Status.State.String(),
		},
		Weather: struct {
			DisplayValue string `json:"display_value"`
			Temperature  int    `json:"temperature"`
		}{
			DisplayValue: g.Weather.DisplayValue,
			Temperature:  g.Weather.Temperature,
		},
		HomeTeam: pubsub.TeamScore{
			Score:            g.HomeTeam.Score,
			Name:             g.HomeTeam.Name,
			ShortDisplayName: g.HomeTeam.ShortDisplayName,
			Logo:             g.HomeTeam.Logo,
			Record:           g.HomeTeam.Record,
		},
		AwayTeam: pubsub.TeamScore{
			Score:            g.AwayTeam.Score,
			Name:             g.AwayTeam.Name,
			ShortDisplayName: g.AwayTeam.ShortDisplayName,
			Logo:             g.AwayTeam.Logo,
			Record:           g.AwayTeam.Record,
		},
		WeekName:       g.WeekName,
		Competition:    g.Competition.String(),
		LastGameChange: lastGameChange.String(),
	})

	return mb
}