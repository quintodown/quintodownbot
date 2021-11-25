package games_test

import (
	"errors"
	"testing"
	"time"

	"github.com/quintodown/quintodownbot/mocks/clock"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/quintodown/quintodownbot/internal/games"
	"github.com/quintodown/quintodownbot/internal/pubsub"
	mgms "github.com/quintodown/quintodownbot/mocks/games"
	mps "github.com/quintodown/quintodownbot/mocks/pubsub"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGameHandler_GetGamesEmptyList(t *testing.T) {
	t.Run("it should return an empty list when games shouldn't be initialised", func(t *testing.T) {
		gh := games.NewGameHandler(new(mgms.GameInfoClient), false, new(mps.Queue), new(clock.Clock))

		require.Empty(t, gh.GetGames(games.NFL))
	})

	t.Run("it should return an empty list when failed initialising games", func(t *testing.T) {
		gic := new(mgms.GameInfoClient)
		initialised := make(chan interface{})

		gic.On("GetGames", games.NFL).Once().Return(func(games.Competition) []games.Game {
			defer close(initialised)

			return nil
		}, errors.New("failing"))

		gh := games.NewGameHandler(gic, true, new(mps.Queue), new(clock.Clock))

		<-initialised

		require.Eventually(t, func() bool { return len(gic.Calls) > 0 }, time.Second, time.Millisecond)
		require.Empty(t, gh.GetGames(games.NFL))
		gic.AssertExpectations(t)
	})

	t.Run("it should an empty list when all games started more than 4 days ago", func(t *testing.T) {
		gic := new(mgms.GameInfoClient)
		mclk := new(clock.Clock)
		initialised := make(chan interface{})
		now := time.Now().UTC()

		gic.On("GetGames", games.NFL).Once().Return(func(games.Competition) []games.Game {
			defer close(initialised)

			return []games.Game{
				{
					Id:          "qwert",
					Start:       time.Now().UTC().Add(-5 * 24 * time.Hour),
					Competition: games.NFL,
				},
			}
		}, nil)
		mclk.On("Now").Once().Return(now)

		gh := games.NewGameHandler(gic, true, new(mps.Queue), mclk)

		<-initialised
		require.Eventually(t, func() bool { return len(gic.Calls) > 0 }, time.Second, time.Millisecond)

		require.Empty(t, gh.GetGames(games.NFL))
		gic.AssertExpectations(t)
	})
}

func TestGameHandler_GetGames(t *testing.T) {
	t.Run("it should return a list of games sorted by start date after they have been initialised", func(t *testing.T) {
		gic := new(mgms.GameInfoClient)
		mclk := new(clock.Clock)
		initialised := make(chan interface{})
		now := time.Now().UTC()

		gic.On("GetGames", games.NFL).Once().Return(func(games.Competition) []games.Game {
			defer close(initialised)

			return []games.Game{
				{
					Id:    "asdfg",
					Start: now.Add(3 * time.Hour),
				},
				{
					Id:    "gfdsa",
					Start: now.Add(2 * time.Hour),
				},
			}
		}, nil)
		mclk.On("Now").Once().Return(now)

		gh := games.NewGameHandler(gic, true, new(mps.Queue), mclk)

		<-initialised
		require.Eventually(t, func() bool { return len(gic.Calls) > 0 }, time.Second, time.Millisecond)

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
	mclk := new(clock.Clock)

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
	mclk.On("Now").Times(2).Return(now)

	gh := games.NewGameHandler(gic, true, q, mclk)

	<-initialised
	require.Eventually(t, func() bool { return len(gic.Calls) > 0 }, time.Second, time.Millisecond)

	getGames := gh.GetGamesStartingIn(games.NFL, time.Hour)
	require.Len(t, getGames, 1)
	require.Equal(t, "gfdsa", getGames[0].Id)

	gic.AssertExpectations(t)
	mclk.AssertExpectations(t)
}

func TestGameHandler_GetGame(t *testing.T) {
	gic := new(mgms.GameInfoClient)
	q := new(mps.Queue)
	mclk := new(clock.Clock)
	now := time.Now().UTC()

	initialised := make(chan interface{})
	g1 := games.Game{
		Id:    "asdfg",
		Start: now.Add(5 * time.Hour),
	}

	gic.On("GetGames", games.NFL).Once().Return(func(games.Competition) []games.Game {
		defer close(initialised)

		return []games.Game{
			g1,
			{
				Id:    "gfdsa",
				Start: now.Add(time.Hour),
			},
		}
	}, nil)
	mclk.On("Now").Once().Return(now)

	gh := games.NewGameHandler(gic, true, q, mclk)

	<-initialised
	require.Eventually(t, func() bool { return len(gh.GetGames(games.NFL)) > 0 }, time.Second, time.Millisecond)

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
		gic, _, gh := initGameHandler(t, startPlaying, games.Game{}, errors.New("testing"), "")

		gh.UpdateGamesInformation(true)
		mockAssertion(t, gic, nil)
	})

	t.Run("it should not send any update when there isn't any update", func(t *testing.T) {
		gic, _, gh := initGameHandler(t, startPlaying, gameInProgress(startPlaying), nil, "")

		gh.UpdateGamesInformation(true)
		mockAssertion(t, gic, nil)
	})

	t.Run("it should send game information when game has been rescheduled", func(t *testing.T) {
		newTime := time.Now().UTC().Add(4 * time.Hour)

		gic, q, gh := initGameHandler(t, startPlaying, gameRescheduled(newTime), nil, rescheduledPayload(newTime))

		gh.UpdateGamesInformation(true)
		mockAssertion(t, gic, q)
	})

	t.Run("it should send game information when home team scores", func(t *testing.T) {
		gic, q, gh := initGameHandler(t, startPlaying, gameHomeScore(startPlaying), nil, homeScorePayload(startPlaying))

		gh.UpdateGamesInformation(true)
		mockAssertion(t, gic, q)
	})

	t.Run("it should send game information when away team scores", func(t *testing.T) {
		gic, q, gh := initGameHandler(t, startPlaying, gameAwayScore(startPlaying), nil, awayScorePayload(startPlaying))

		gh.UpdateGamesInformation(true)
		mockAssertion(t, gic, q)
	})

	t.Run("it should send game information when game has started", func(t *testing.T) {
		gic, q, gh := initGameHandler(t, startPlaying, gameStarted(startPlaying), nil, getStartedPayload(startPlaying))

		gh.UpdateGamesInformation(true)
		mockAssertion(t, gic, q)
	})

	t.Run("it should send game information when period has finished", func(t *testing.T) {
		gic, q, gh := initGameHandler(t, startPlaying, gameEndPeriod(startPlaying), nil, endPeriodPayload(startPlaying))

		gh.UpdateGamesInformation(true)
		mockAssertion(t, gic, q)
	})

	t.Run("it should send game information when game has finished", func(t *testing.T) {
		gic, q, gh := initGameHandler(t, startPlaying, gameFinished(startPlaying), nil, gameFinishedPayload(startPlaying))

		gh.UpdateGamesInformation(true)
		mockAssertion(t, gic, q)
	})
}

func mockAssertion(t *testing.T, gic *mgms.GameInfoClient, q *mps.Queue) {
	gic.AssertExpectations(t)

	if q != nil {
		q.AssertExpectations(t)
	}
}

func initGameHandler(t *testing.T, startPlaying time.Time, g games.Game, e error, payload string) (
	*mgms.GameInfoClient,
	*mps.Queue,
	games.Handler,
) {
	gic := new(mgms.GameInfoClient)
	q := new(mps.Queue)
	mclk := new(clock.Clock)

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
	gic.On("GetGameInformation", games.NFL, "asdfg").Once().Return(g, e)

	mclk.On("Now").Return(time.Now().UTC())

	gh := games.NewGameHandler(gic, true, q, mclk)

	if len(payload) > 0 {
		q.On(
			"Publish",
			pubsub.GamesTopic.String(),
			mock.MatchedBy(func(m *message.Message) bool {
				return string(m.Payload) == payload
			}),
		).Once().Return(nil)
	}

	<-initialised

	require.Eventually(t, func() bool { return len(gh.GetGames(games.NFL)) == 2 }, time.Second, time.Millisecond)

	return gic, q, gh
}

func rescheduledPayload(newTime time.Time) string {
	return "{\"id\":\"asdfg\"," +
		"\"start\":\"" + newTime.Format(time.RFC3339Nano) + "\"," +
		"\"name\":\"\",\"venue\":{\"fullName\":\"\",\"city\":\"\"," +
		"\"state\":\"\",\"capacity\":0,\"indoor\":false}," +
		"\"status\":{\"clock\":0,\"displayClock\":\"\",\"period\":0," +
		"\"state\":\"RescheduledState\"},\"weather\":{\"displayValue\":\"\"," +
		"\"temperature\":0},\"homeTeam\":{\"score\":0,\"name\":\"\"," +
		"\"shortDisplayName\":\"\",\"logo\":\"\",\"record\":\"\"}," +
		"\"awayTeam\":{\"score\":0,\"name\":\"\",\"shortDisplayName\":\"\"," +
		"\"logo\":\"\",\"record\":\"\"},\"weekName\":\"\"," +
		"\"competition\":\"NFL\",\"lastGameChange\":\"Rescheduled\"}"
}

func gameFinishedPayload(startPlaying time.Time) string {
	return "{\"id\":\"asdfg\"," +
		"\"start\":\"" + startPlaying.Format(time.RFC3339Nano) + "\",\"name\":\"\"," +
		"\"venue\":{\"fullName\":\"\",\"city\":\"\",\"state\":\"\"," +
		"\"capacity\":0,\"indoor\":false}," +
		"\"status\":{\"clock\":0,\"displayClock\":\"\",\"period\":0," +
		"\"state\":\"FinishedState\"}," +
		"\"weather\":{\"displayValue\":\"\",\"temperature\":0}," +
		"\"homeTeam\":{\"score\":0,\"name\":\"\"," +
		"\"shortDisplayName\":\"\",\"logo\":\"\",\"record\":\"\"}," +
		"\"awayTeam\":{\"score\":0,\"name\":\"\"," +
		"\"shortDisplayName\":\"\",\"logo\":\"\",\"record\":\"\"}," +
		"\"weekName\":\"\",\"competition\":\"NFL\"," +
		"\"lastGameChange\":\"Finished\"}"
}

func endPeriodPayload(startPlaying time.Time) string {
	return "{\"id\":\"asdfg\"," +
		"\"start\":\"" + startPlaying.Format(time.RFC3339Nano) + "\",\"name\":\"\"," +
		"\"venue\":{\"fullName\":\"\",\"city\":\"\",\"state\":\"\"," +
		"\"capacity\":0,\"indoor\":false}," +
		"\"status\":{\"clock\":0,\"displayClock\":\"\",\"period\":2," +
		"\"state\":\"InProgressState\"}," +
		"\"weather\":{\"displayValue\":\"\",\"temperature\":0}," +
		"\"homeTeam\":{\"score\":0,\"name\":\"\"," +
		"\"shortDisplayName\":\"\",\"logo\":\"\",\"record\":\"\"}," +
		"\"awayTeam\":{\"score\":0,\"name\":\"\"," +
		"\"shortDisplayName\":\"\",\"logo\":\"\",\"record\":\"\"}," +
		"\"weekName\":\"\",\"competition\":\"NFL\"," +
		"\"lastGameChange\":\"PeriodFinished\"}"
}

func getStartedPayload(startPlaying time.Time) string {
	return "{\"id\":\"asdfg\"," +
		"\"start\":\"" + startPlaying.Format(time.RFC3339Nano) + "\",\"name\":\"\"," +
		"\"venue\":{\"fullName\":\"\",\"city\":\"\",\"state\":\"\"," +
		"\"capacity\":0,\"indoor\":false}," +
		"\"status\":{\"clock\":0,\"displayClock\":\"\",\"period\":1," +
		"\"state\":\"InProgressState\"}," +
		"\"weather\":{\"displayValue\":\"\",\"temperature\":0}," +
		"\"homeTeam\":{\"score\":0,\"name\":\"\"," +
		"\"shortDisplayName\":\"\",\"logo\":\"\",\"record\":\"\"}," +
		"\"awayTeam\":{\"score\":0,\"name\":\"\"," +
		"\"shortDisplayName\":\"\",\"logo\":\"\",\"record\":\"\"}," +
		"\"weekName\":\"\",\"competition\":\"NFL\"," +
		"\"lastGameChange\":\"Started\"}"
}

func awayScorePayload(startPlaying time.Time) string {
	return "{\"id\":\"asdfg\"," +
		"\"start\":\"" + startPlaying.Format(time.RFC3339Nano) + "\",\"name\":\"\"," +
		"\"venue\":{\"fullName\":\"\",\"city\":\"\",\"state\":\"\"," +
		"\"capacity\":0,\"indoor\":false}," +
		"\"status\":{\"clock\":0,\"displayClock\":\"\",\"period\":0," +
		"\"state\":\"InProgressState\"}," +
		"\"weather\":{\"displayValue\":\"\",\"temperature\":0}," +
		"\"homeTeam\":{\"score\":0,\"name\":\"\"," +
		"\"shortDisplayName\":\"\",\"logo\":\"\",\"record\":\"\"}," +
		"\"awayTeam\":{\"score\":7,\"name\":\"\"," +
		"\"shortDisplayName\":\"\",\"logo\":\"\",\"record\":\"\"}," +
		"\"weekName\":\"\",\"competition\":\"NFL\"," +
		"\"lastGameChange\":\"AwayScore\"}"
}

func homeScorePayload(startPlaying time.Time) string {
	return "{\"id\":\"asdfg\"," +
		"\"start\":\"" + startPlaying.Format(time.RFC3339Nano) + "\",\"name\":\"\"," +
		"\"venue\":{\"fullName\":\"\",\"city\":\"\",\"state\":\"\"," +
		"\"capacity\":0,\"indoor\":false}," +
		"\"status\":{\"clock\":0,\"displayClock\":\"\",\"period\":0," +
		"\"state\":\"InProgressState\"}," +
		"\"weather\":{\"displayValue\":\"\",\"temperature\":0}," +
		"\"homeTeam\":{\"score\":7,\"name\":\"\"," +
		"\"shortDisplayName\":\"\",\"logo\":\"\",\"record\":\"\"}," +
		"\"awayTeam\":{\"score\":0,\"name\":\"\"," +
		"\"shortDisplayName\":\"\",\"logo\":\"\",\"record\":\"\"}," +
		"\"weekName\":\"\",\"competition\":\"NFL\"," +
		"\"lastGameChange\":\"HomeScore\"}"
}

func gameFinished(startPlaying time.Time) games.Game {
	return games.Game{
		Id:     "asdfg",
		Start:  startPlaying,
		Status: games.GameStatus{State: games.FinishedState},
	}
}

func gameEndPeriod(startPlaying time.Time) games.Game {
	return games.Game{
		Id:    "asdfg",
		Start: startPlaying,
		Status: games.GameStatus{
			Period: 2,
			State:  games.InProgressState,
		},
	}
}

func gameStarted(startPlaying time.Time) games.Game {
	return games.Game{
		Id:    "asdfg",
		Start: startPlaying,
		Status: games.GameStatus{
			Period: 1,
			State:  games.InProgressState,
		},
	}
}

func gameAwayScore(startPlaying time.Time) games.Game {
	return games.Game{
		Id:       "asdfg",
		Start:    startPlaying,
		AwayTeam: games.TeamScore{Score: 7},
	}
}

func gameHomeScore(startPlaying time.Time) games.Game {
	return games.Game{
		Id:       "asdfg",
		Start:    startPlaying,
		HomeTeam: games.TeamScore{Score: 7},
	}
}

func gameRescheduled(newTime time.Time) games.Game {
	return games.Game{
		Id:    "asdfg",
		Start: newTime,
	}
}

func gameInProgress(startPlaying time.Time) games.Game {
	return games.Game{
		Id:          "asdfg",
		Start:       startPlaying,
		Status:      games.GameStatus{State: games.InProgressState},
		Competition: games.NFL,
	}
}
