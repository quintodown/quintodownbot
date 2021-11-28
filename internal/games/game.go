package games

import (
	"errors"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/quintodown/quintodownbot/internal/clock"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/mailru/easyjson"
	"github.com/quintodown/quintodownbot/internal/pubsub"
)

const timeToCleanup = 4 * 24 * time.Hour

type GameInfoClient interface {
	GetGames(Competition) ([]Game, error)
	GetGameInformation(Competition, string) (Game, error)
}

type Handler interface {
	GetGames(c Competition) []Game
	GetGamesStartingIn(c Competition, d time.Duration) []Game
	GetGame(id string) (Game, error)
	UpdateGamesInformation(onlyPlaying bool)
	UpdateGamesList()
}

type GameHandler struct {
	client   GameInfoClient
	gameList sync.Map
	queue    pubsub.Queue
	clk      clock.Clock
}

func NewGameHandler(client GameInfoClient, getGames bool, queue pubsub.Queue, clk clock.Clock) Handler {
	gh := &GameHandler{client: client, queue: queue, clk: clk}

	if getGames {
		go func() { gh.UpdateGamesList() }()
	}

	return gh
}

func (gh *GameHandler) GetGames(c Competition) []Game {
	list := gh.gamesList(c)
	found := make([]Game, 0, len(list))

	for k := range list {
		found = append(found, list[k])
	}

	sort.Slice(found, func(i, j int) bool {
		return found[i].Start.Before(found[j].Start)
	})

	return found
}

func (gh *GameHandler) GetGamesStartingIn(c Competition, d time.Duration) []Game {
	var found []Game

	offset := gh.clk.Now().Add(d).Truncate(time.Minute)

	for _, v := range gh.gamesList(c) {
		if offset.Equal(v.Start.UTC().Truncate(time.Minute)) {
			found = append(found, v)
		}
	}

	return found
}

func (gh *GameHandler) GetGame(id string) (Game, error) {
	for _, competition := range GetCompetitions() {
		for _, v := range gh.gamesList(competition) {
			if v.Id == id {
				return v, nil
			}
		}
	}

	return Game{}, errors.New("game not found")
}

func (gh *GameHandler) UpdateGamesInformation(onlyPlaying bool) {
	for _, competition := range GetCompetitions() {
		gameList := gh.gamesList(competition)
		for k, v := range gameList {
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
				v.Start = g.Start
				v.Status.State = RescheduledState
			case HomeScore:
				v.HomeTeam.Score = g.HomeTeam.Score
			case AwayScore:
				v.AwayTeam.Score = g.AwayTeam.Score
			case Started:
				v.Status.State = InProgressState
				v.Status.Period = 1
				v.Status.DisplayClock = g.Status.DisplayClock
			case PeriodFinished:
				v.Status.Period = g.Status.Period
				v.Status.DisplayClock = g.Status.DisplayClock
			case Finished:
				v.Status = g.Status
				v.HomeTeam.Score = g.HomeTeam.Score
				v.AwayTeam.Score = g.AwayTeam.Score
				v.Status.Period = g.Status.Period
				v.Status.DisplayClock = g.Status.DisplayClock
			}

			if lastGameChange != NoChanges {
				gameList[k] = v
				_ = gh.queue.Publish(
					pubsub.GamesTopic.String(),
					message.NewMessage(
						watermill.NewUUID(),
						v.toGameEvent(lastGameChange),
					),
				)
			}
		}

		gh.gameList.Store(competition.String(), gameList)
	}
}

func (gh *GameHandler) UpdateGamesList() {
	for _, v := range GetCompetitions() {
		list := gh.gamesList(v)

		games, err := gh.client.GetGames(v)
		if err != nil {
			continue
		}

		for i := range games {
			if _, ok := list[games[i].key()]; !ok {
				list[games[i].key()] = games[i]
			}
		}

		gh.gameList.Store(v.String(), gh.cleanUpGames(list))
	}
}

func (gh *GameHandler) gamesList(c Competition) map[string]Game {
	f, ok := gh.gameList.Load(c.String())
	if !ok {
		return map[string]Game{}
	}

	games, ok := f.(map[string]Game)
	if !ok {
		return nil
	}

	return games
}

func (gh *GameHandler) getLastGameChange(oldGameInfo, newGameInfo Game) GameChange {
	lastGameChange := NoChanges

	if reflect.DeepEqual(oldGameInfo, newGameInfo) {
		return lastGameChange
	}

	if !newGameInfo.Start.Equal(oldGameInfo.Start) && oldGameInfo.Status.State != RescheduledState {
		lastGameChange = Rescheduled
	}

	if lastGameChange == NoChanges && newGameInfo.HomeTeam.Score != oldGameInfo.HomeTeam.Score {
		lastGameChange = HomeScore
	}

	if lastGameChange == NoChanges && newGameInfo.AwayTeam.Score != oldGameInfo.AwayTeam.Score {
		lastGameChange = AwayScore
	}

	if lastGameChange == NoChanges &&
		newGameInfo.Status.State == FinishedState && oldGameInfo.Status.State != FinishedState {
		lastGameChange = Finished
	}

	if lastGameChange == NoChanges &&
		(newGameInfo.Status.State != oldGameInfo.Status.State || newGameInfo.Status.Period != oldGameInfo.Status.Period) {
		if newGameInfo.Status.Period == 1 {
			lastGameChange = Started
		} else {
			lastGameChange = PeriodFinished
		}
	}

	return lastGameChange
}

func (gh *GameHandler) cleanUpGames(gms map[string]Game) map[string]Game {
	now := gh.clk.Now()
	for k := range gms {
		if now.Sub(gms[k].Start) > timeToCleanup {
			delete(gms, k)
		}
	}

	return gms
}

func (g *Game) toGameEvent(lastGameChange GameChange) []byte {
	mb, _ := easyjson.Marshal(pubsub.GameEvent{
		Id:    g.Id,
		Start: g.Start,
		Name:  g.Name,
		Venue: pubsub.GameVenue{
			FullName: g.Venue.FullName,
			City:     g.Venue.Address.City,
			State:    g.Venue.Address.State,
			Capacity: g.Venue.Capacity,
			Indoor:   g.Venue.Indoor,
		},
		Status: pubsub.GameStatus{
			Clock:        g.Status.Clock,
			DisplayClock: g.Status.DisplayClock,
			Period:       g.Status.Period,
			State:        g.Status.State.String(),
		},
		Weather: pubsub.GameWeather{
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
