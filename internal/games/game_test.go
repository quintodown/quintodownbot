package games_test

import (
	"errors"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/quintodown/quintodownbot/internal/games"
	"github.com/quintodown/quintodownbot/internal/pubsub"
	mapp "github.com/quintodown/quintodownbot/mocks/app"
	mgms "github.com/quintodown/quintodownbot/mocks/games"
	mps "github.com/quintodown/quintodownbot/mocks/pubsub"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGameHandler_GetGames(t *testing.T) {
	t.Run("it should return an empty list when games shouldn't be initialised", func(t *testing.T) {
		gic := new(mgms.GameInfoClient)
		q := new(mps.Queue)
		mclk := new(mapp.Clock)

		gh := games.NewGameHandler(gic, false, q, mclk)

		require.Empty(t, gh.GetGames(games.NFL))
	})

	t.Run("it should return an empty list when failed initialising games", func(t *testing.T) {
		gic := new(mgms.GameInfoClient)
		q := new(mps.Queue)
		mclk := new(mapp.Clock)

		initialised := make(chan interface{})
		gic.On("GetGames", games.NFL).Once().Return(func(games.Competition) []games.Game {
			defer close(initialised)

			return nil
		}, errors.New("failing"))

		gh := games.NewGameHandler(gic, true, q, mclk)

		<-initialised

		require.Empty(t, gh.GetGames(games.NFL))
		gic.AssertExpectations(t)
	})

	t.Run("it should return a list of games sorted by start date after they have been initialised", func(t *testing.T) {
		gic := new(mgms.GameInfoClient)
		q := new(mps.Queue)
		mclk := new(mapp.Clock)

		initialised := make(chan interface{})

		gic.On("GetGames", games.NFL).Once().Return(func(games.Competition) []games.Game {
			defer close(initialised)

			return []games.Game{
				{
					Id:    "asdfg",
					Start: time.Now().UTC().Add(3 * time.Hour),
				},
				{
					Id:    "gfdsa",
					Start: time.Now().UTC().Add(2 * time.Hour),
				},
			}
		}, nil)

		gh := games.NewGameHandler(gic, true, q, mclk)

		<-initialised

		getGames := gh.GetGames(games.NFL)

		require.Len(t, getGames, 2)
		require.Condition(t, func() bool {
			return getGames[0].Start.Before(getGames[1].Start)
		})
		gic.AssertExpectations(t)
	})
}

func TestGameHandler_GetGamesStartingIn(t *testing.T) {
	gic := new(mgms.GameInfoClient)
	q := new(mps.Queue)
	mclk := new(mapp.Clock)

	initialised := make(chan interface{})
	now := time.Now().UTC()
	gic.On("GetGames", games.NFL).Once().Return(func(games.Competition) []games.Game {
		defer close(initialised)
		return []games.Game{
			{
				Id:    "asdfg",
				Start: now.Add(5 * time.Hour),
			},
			{
				Id:    "gfdsa",
				Start: now.Add(time.Hour),
			},
		}
	}, nil)
	mclk.On("Now").Once().Return(now)

	gh := games.NewGameHandler(gic, true, q, mclk)

	<-initialised

	getGames := gh.GetGamesStartingIn(games.NFL, time.Hour)
	require.Len(t, getGames, 1)
	require.Equal(t, "gfdsa", getGames[0].Id)

	gic.AssertExpectations(t)
	mclk.AssertExpectations(t)
}

func TestGameHandler_GetGame(t *testing.T) {
	gic := new(mgms.GameInfoClient)
	q := new(mps.Queue)
	mclk := new(mapp.Clock)

	initialised := make(chan interface{})
	g1 := games.Game{
		Id:    "asdfg",
		Start: time.Now().UTC().Add(5 * time.Hour),
	}
	gic.On("GetGames", games.NFL).Once().Return(func(games.Competition) []games.Game {
		defer close(initialised)
		return []games.Game{
			g1,
			{
				Id:    "gfdsa",
				Start: time.Now().UTC().Add(time.Hour),
			},
		}
	}, nil)

	gh := games.NewGameHandler(gic, true, q, mclk)

	<-initialised

	t.Run("it should return requested game", func(t *testing.T) {
		game, err := gh.GetGame("asdfg")

		require.NoError(t, err)
		require.Equal(t, g1, game)

		gic.AssertExpectations(t)
	})

	t.Run("it should fail when requested game not found", func(t *testing.T) {
		_, err := gh.GetGame("12345")

		require.EqualError(t, err, "game not found")
		gic.AssertExpectations(t)
	})
}

func TestGameHandler_UpdateGamesInformation(t *testing.T) {
	startPlaying := time.Now().UTC().Add(-1 * time.Hour)

	t.Run("it should not send any update when fails getting game update", func(t *testing.T) {
		gic, _, gh := initialiseGameHandler(t, startPlaying)

		gic.On("GetGameInformation", games.NFL, "asdfg").
			Once().
			Return(games.Game{}, errors.New("testing"))

		gh.UpdateGamesInformation(true)

		gic.AssertExpectations(t)
		gic = nil
		gh = nil
	})

	t.Run("it should send game information when game has been rescheduled", func(t *testing.T) {
		gic, q, gh := initialiseGameHandler(t, startPlaying)

		newTime := time.Now().UTC().Add(4 * time.Hour)

		gic.On("GetGameInformation", games.NFL, "asdfg").
			Once().
			Return(games.Game{
				Id:    "asdfg",
				Start: newTime,
			}, nil)
		q.On(
			"Publish",
			pubsub.GamesTopic.String(),
			mock.MatchedBy(func(m *message.Message) bool {
				return string(m.Payload) == "{\"id\":\"asdfg\","+
					"\"start\":\""+newTime.Format(time.RFC3339Nano)+"\","+
					"\"name\":\"\",\"venue\":{\"full_name\":\"\",\"city\":\"\","+
					"\"state\":\"\",\"capacity\":0,\"indoor\":false},"+
					"\"status\":{\"clock\":0,\"display_clock\":\"\",\"period\":0,"+
					"\"state\":\"RescheduledState\"},\"weather\":{\"display_value\":\"\","+
					"\"temperature\":0},\"home_team\":{\"score\":0,\"name\":\"\","+
					"\"short_display_name\":\"\",\"logo\":\"\",\"record\":\"\"},"+
					"\"away_team\":{\"score\":0,\"name\":\"\",\"short_display_name\":\"\","+
					"\"logo\":\"\",\"record\":\"\"},\"week_name\":\"\","+
					"\"competition\":\"NFL\",\"last_game_change\":\"Rescheduled\"}"
			}),
		).Once().Return(nil)

		gh.UpdateGamesInformation(true)

		gic.AssertExpectations(t)
		q.AssertExpectations(t)
		gic = nil
		q = nil
		gh = nil
	})

	t.Run("it should send game information when home team scores", func(t *testing.T) {
		gic, q, gh := initialiseGameHandler(t, startPlaying)

		gic.On("GetGameInformation", games.NFL, "asdfg").
			Once().
			Return(games.Game{
				Id:       "asdfg",
				Start:    startPlaying,
				HomeTeam: games.TeamScore{Score: 7},
			}, nil)
		q.On(
			"Publish",
			pubsub.GamesTopic.String(),
			mock.MatchedBy(func(m *message.Message) bool {
				return string(m.Payload) == "{\"id\":\"asdfg\","+
					"\"start\":\""+startPlaying.Format(time.RFC3339Nano)+"\",\"name\":\"\","+
					"\"venue\":{\"full_name\":\"\",\"city\":\"\",\"state\":\"\","+
					"\"capacity\":0,\"indoor\":false},"+
					"\"status\":{\"clock\":0,\"display_clock\":\"\",\"period\":0,"+
					"\"state\":\"InProgressState\"},"+
					"\"weather\":{\"display_value\":\"\",\"temperature\":0},"+
					"\"home_team\":{\"score\":7,\"name\":\"\","+
					"\"short_display_name\":\"\",\"logo\":\"\",\"record\":\"\"},"+
					"\"away_team\":{\"score\":0,\"name\":\"\","+
					"\"short_display_name\":\"\",\"logo\":\"\",\"record\":\"\"},"+
					"\"week_name\":\"\",\"competition\":\"NFL\","+
					"\"last_game_change\":\"HomeScore\"}"
			}),
		).Once().Return(nil)

		gh.UpdateGamesInformation(true)

		gic.AssertExpectations(t)
		q.AssertExpectations(t)
		gic = nil
		q = nil
		gh = nil
	})

	t.Run("it should send game information when away team scores", func(t *testing.T) {
		gic, q, gh := initialiseGameHandler(t, startPlaying)

		gic.On("GetGameInformation", games.NFL, "asdfg").
			Once().
			Return(games.Game{
				Id:       "asdfg",
				Start:    startPlaying,
				AwayTeam: games.TeamScore{Score: 7},
			}, nil)
		q.On(
			"Publish",
			pubsub.GamesTopic.String(),
			mock.MatchedBy(func(m *message.Message) bool {
				return string(m.Payload) == "{\"id\":\"asdfg\","+
					"\"start\":\""+startPlaying.Format(time.RFC3339Nano)+"\",\"name\":\"\","+
					"\"venue\":{\"full_name\":\"\",\"city\":\"\",\"state\":\"\","+
					"\"capacity\":0,\"indoor\":false},"+
					"\"status\":{\"clock\":0,\"display_clock\":\"\",\"period\":0,"+
					"\"state\":\"InProgressState\"},"+
					"\"weather\":{\"display_value\":\"\",\"temperature\":0},"+
					"\"home_team\":{\"score\":0,\"name\":\"\","+
					"\"short_display_name\":\"\",\"logo\":\"\",\"record\":\"\"},"+
					"\"away_team\":{\"score\":7,\"name\":\"\","+
					"\"short_display_name\":\"\",\"logo\":\"\",\"record\":\"\"},"+
					"\"week_name\":\"\",\"competition\":\"NFL\","+
					"\"last_game_change\":\"AwayScore\"}"
			}),
		).Once().Return(nil)

		gh.UpdateGamesInformation(true)

		gic.AssertExpectations(t)
		q.AssertExpectations(t)
		gic = nil
		q = nil
		gh = nil
	})

	t.Run("it should send game information when game has started", func(t *testing.T) {
		gic, q, gh := initialiseGameHandler(t, startPlaying)

		gic.On("GetGameInformation", games.NFL, "asdfg").
			Once().
			Return(games.Game{
				Id:    "asdfg",
				Start: startPlaying,
				Status: games.GameStatus{
					Period: 1,
					State:  games.InProgressState,
				},
			}, nil)
		q.On(
			"Publish",
			pubsub.GamesTopic.String(),
			mock.MatchedBy(func(m *message.Message) bool {
				return string(m.Payload) == "{\"id\":\"asdfg\","+
					"\"start\":\""+startPlaying.Format(time.RFC3339Nano)+"\",\"name\":\"\","+
					"\"venue\":{\"full_name\":\"\",\"city\":\"\",\"state\":\"\","+
					"\"capacity\":0,\"indoor\":false},"+
					"\"status\":{\"clock\":0,\"display_clock\":\"\",\"period\":1,"+
					"\"state\":\"InProgressState\"},"+
					"\"weather\":{\"display_value\":\"\",\"temperature\":0},"+
					"\"home_team\":{\"score\":0,\"name\":\"\","+
					"\"short_display_name\":\"\",\"logo\":\"\",\"record\":\"\"},"+
					"\"away_team\":{\"score\":0,\"name\":\"\","+
					"\"short_display_name\":\"\",\"logo\":\"\",\"record\":\"\"},"+
					"\"week_name\":\"\",\"competition\":\"NFL\","+
					"\"last_game_change\":\"Started\"}"
			}),
		).Once().Return(nil)

		gh.UpdateGamesInformation(true)

		gic.AssertExpectations(t)
		q.AssertExpectations(t)
		gic = nil
		q = nil
		gh = nil
	})

	t.Run("it should send game information when period has finished", func(t *testing.T) {
		gic, q, gh := initialiseGameHandler(t, startPlaying)

		gic.On("GetGameInformation", games.NFL, "asdfg").
			Once().
			Return(games.Game{
				Id:    "asdfg",
				Start: startPlaying,
				Status: games.GameStatus{
					Period: 2,
					State:  games.InProgressState,
				},
			}, nil)
		q.On(
			"Publish",
			pubsub.GamesTopic.String(),
			mock.MatchedBy(func(m *message.Message) bool {
				return string(m.Payload) == "{\"id\":\"asdfg\","+
					"\"start\":\""+startPlaying.Format(time.RFC3339Nano)+"\",\"name\":\"\","+
					"\"venue\":{\"full_name\":\"\",\"city\":\"\",\"state\":\"\","+
					"\"capacity\":0,\"indoor\":false},"+
					"\"status\":{\"clock\":0,\"display_clock\":\"\",\"period\":2,"+
					"\"state\":\"InProgressState\"},"+
					"\"weather\":{\"display_value\":\"\",\"temperature\":0},"+
					"\"home_team\":{\"score\":0,\"name\":\"\","+
					"\"short_display_name\":\"\",\"logo\":\"\",\"record\":\"\"},"+
					"\"away_team\":{\"score\":0,\"name\":\"\","+
					"\"short_display_name\":\"\",\"logo\":\"\",\"record\":\"\"},"+
					"\"week_name\":\"\",\"competition\":\"NFL\","+
					"\"last_game_change\":\"PeriodFinished\"}"
			}),
		).Once().Return(nil)

		gh.UpdateGamesInformation(true)

		gic.AssertExpectations(t)
		q.AssertExpectations(t)
		gic = nil
		q = nil
		gh = nil
	})

	t.Run("it should send game information when game has finished", func(t *testing.T) {
		gic, q, gh := initialiseGameHandler(t, startPlaying)

		gic.On("GetGameInformation", games.NFL, "asdfg").
			Once().
			Return(games.Game{
				Id:     "asdfg",
				Start:  startPlaying,
				Status: games.GameStatus{State: games.FinishedState},
			}, nil)
		q.On(
			"Publish",
			pubsub.GamesTopic.String(),
			mock.MatchedBy(func(m *message.Message) bool {
				return string(m.Payload) == "{\"id\":\"asdfg\","+
					"\"start\":\""+startPlaying.Format(time.RFC3339Nano)+"\",\"name\":\"\","+
					"\"venue\":{\"full_name\":\"\",\"city\":\"\",\"state\":\"\","+
					"\"capacity\":0,\"indoor\":false},"+
					"\"status\":{\"clock\":0,\"display_clock\":\"\",\"period\":0,"+
					"\"state\":\"FinishedState\"},"+
					"\"weather\":{\"display_value\":\"\",\"temperature\":0},"+
					"\"home_team\":{\"score\":0,\"name\":\"\","+
					"\"short_display_name\":\"\",\"logo\":\"\",\"record\":\"\"},"+
					"\"away_team\":{\"score\":0,\"name\":\"\","+
					"\"short_display_name\":\"\",\"logo\":\"\",\"record\":\"\"},"+
					"\"week_name\":\"\",\"competition\":\"NFL\","+
					"\"last_game_change\":\"GameFinished\"}"
			}),
		).Once().Return(nil)

		gh.UpdateGamesInformation(true)

		gic.AssertExpectations(t)
		q.AssertExpectations(t)
		gic = nil
		q = nil
		gh = nil
	})

}

func initialiseGameHandler(t *testing.T, startPlaying time.Time) (*mgms.GameInfoClient, *mps.Queue, *games.GameHandler) {
	gic := new(mgms.GameInfoClient)
	q := new(mps.Queue)
	mclk := new(mapp.Clock)

	initialised := make(chan interface{})
	gic.On("GetGames", games.NFL).Once().Return(func(games.Competition) []games.Game {
		defer close(initialised)

		return []games.Game{
			{
				Id:          "asdfg",
				Start:       startPlaying,
				Status:      games.GameStatus{State: games.InProgressState},
				Competition: games.NFL,
			},
			{
				Id:          "gfdsa",
				Start:       time.Now().UTC().Add(time.Hour),
				Competition: games.NFL,
			},
		}
	}, nil)
	mclk.On("Now").Return(time.Now().UTC())

	gh := games.NewGameHandler(gic, true, q, mclk)

	<-initialised

	require.Eventually(t, func() bool { return len(gh.GetGames(games.NFL)) == 2 }, time.Second, time.Millisecond)

	return gic, q, gh
}
