package espn_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/quintodown/quintodownbot/mocks/clock"

	"github.com/jarcoal/httpmock"
	"github.com/quintodown/quintodownbot/internal/games"
	"github.com/quintodown/quintodownbot/internal/games/clients/espn"
	"github.com/stretchr/testify/require"
)

const (
	eventEndpoint           = "https://site.api.espn.com/apis/site/v2/sports/football//summary?event=1&lang=es&region=us"
	scoreboardEndpoint      = "https://site.api.espn.com/apis/site/v2/sports/football//scoreboard?lang=es&region=us"
	scoreboardDatesEndpoint = "https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard?dates=" +
		"20210909-20210916&lang=es&region=us"
)

func TestClient_GetGames(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mclk := new(clock.Clock)

	espnc := espn.NewESPNClient(&http.Client{}, mclk)

	registerMocksHTTP()

	t.Run("it should fail for non existing competition", func(t *testing.T) {
		gms, err := espnc.GetGames(10)

		require.EqualError(t, err, fmt.Sprintf("Get \"%s\": competition not found", scoreboardEndpoint))
		require.Empty(t, gms)
	})

	t.Run("it should fail cause scoreboard not found", func(t *testing.T) {
		gms, err := espnc.GetGames(games.NCAA)

		require.EqualError(t, err, "EOF")
		require.Empty(t, gms)
	})

	t.Run("it should fail cause wrong scoreboard", func(t *testing.T) {
		gms, err := espnc.GetGames(games.CFL)

		require.EqualError(t, err, "parse error: expected { near offset 1 of ''")
		require.Empty(t, gms)
	})

	t.Run("it should fail getting games for current week", func(t *testing.T) {
		mclk.On("Now").
			Once().
			Return(time.Date(2021, 9, 10, 1, 1, 1, 1, time.UTC))

		gms, err := espnc.GetGames(games.NFL)

		require.EqualError(t, err, fmt.Sprintf("Get \"%s\": no responder found", scoreboardDatesEndpoint))
		require.Empty(t, gms)
		mclk.AssertExpectations(t)
	})

	t.Run("it should get games for current week", func(t *testing.T) {
		mclk.On("Now").Once().
			Return(time.Date(2021, 10, 10, 1, 1, 1, 1, time.UTC))

		gms, err := espnc.GetGames(games.NFL)

		marshal, _ := json.Marshal(gms)
		bytes, _ := os.ReadFile("testdata/scoreboard.golden.json")

		require.NoError(t, err)
		require.JSONEq(t, string(bytes), string(marshal))
		mclk.AssertExpectations(t)
	})
}

func TestClient_GetGameInformation(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := &http.Client{}
	espnc := espn.NewESPNClient(client, new(clock.Clock))

	registerMocksHTTP()

	t.Run("it should fail when wrong competition", func(t *testing.T) {
		information, err := espnc.GetGameInformation(10, "1")

		require.EqualError(
			t,
			err,
			fmt.Sprintf("Get \"%s\": no responder found", eventEndpoint),
		)
		require.Empty(t, information)
	})

	t.Run("it should fail when game not found", func(t *testing.T) {
		information, err := espnc.GetGameInformation(games.NFL, "1")

		require.EqualError(t, err, "EOF")
		require.Empty(t, information)
	})

	t.Run("it should fail when couldn't read game information", func(t *testing.T) {
		information, err := espnc.GetGameInformation(games.NFL, "2")

		require.EqualError(t, err, "parse error: expected { near offset 1 of ''")
		require.Empty(t, information)
	})

	t.Run("it should fail when wrong date format", func(t *testing.T) {
		information, err := espnc.GetGameInformation(games.NFL, "4")

		require.EqualError(
			t,
			err,
			"parsing time \"08-19-2021 23:30\" as \"2006-01-02T15:04Z\": cannot parse \"9-2021 23:30\" as \"2006\"",
		)
		require.Empty(t, information)
	})

	t.Run("it should get updated game information", func(t *testing.T) {
		information, err := espnc.GetGameInformation(games.NFL, "3")

		marshal, _ := json.Marshal(information)
		bytes, _ := os.ReadFile("testdata/game.golden.json")

		require.NoError(t, err)
		require.JSONEq(t, string(bytes), string(marshal))
	})
}

func registerMocksHTTP() {
	registerScoreBoardMocks()
	registerGameMocks()
}

func registerScoreBoardMocks() {
	scoreboardResponder := func(req *http.Request) (*http.Response, error) {
		sc, _ := os.ReadFile("testdata/scoreboard.json")

		return httpmock.NewStringResponse(http.StatusOK, string(sc)), nil
	}

	httpmock.RegisterResponder(
		http.MethodGet,
		scoreboardEndpoint,
		httpmock.NewErrorResponder(errors.New("competition not found")),
	)

	httpmock.RegisterResponder(
		http.MethodGet,
		"https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard?lang=es&region=us",
		scoreboardResponder,
	)

	httpmock.RegisterResponder(
		http.MethodGet,
		"https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard?dates=20211007-20211014&lang=es&region=us",
		scoreboardResponder,
	)

	httpmock.RegisterResponder(
		http.MethodGet,
		"https://site.api.espn.com/apis/site/v2/sports/football/cfl/scoreboard?lang=es&region=us",
		httpmock.NewStringResponder(http.StatusOK, "[]"),
	)

	httpmock.RegisterResponder(
		http.MethodGet,
		"https://site.api.espn.com/apis/site/v2/sports/football/college-football/scoreboard?lang=es&region=us",
		httpmock.NewStringResponder(http.StatusNotFound, ""),
	)
}

func registerGameMocks() {
	httpmock.RegisterResponder(
		http.MethodGet,
		"https://site.api.espn.com/apis/site/v2/sports/football/nfl/summary?event=1&lang=es&region=us",
		httpmock.NewStringResponder(http.StatusNotFound, ""),
	)

	httpmock.RegisterResponder(
		http.MethodGet,
		"https://site.api.espn.com/apis/site/v2/sports/football/nfl/summary?event=2&lang=es&region=us",
		httpmock.NewStringResponder(http.StatusOK, "[]"),
	)

	httpmock.RegisterResponder(
		http.MethodGet,
		"https://site.api.espn.com/apis/site/v2/sports/football/nfl/summary?event=3&lang=es&region=us",
		func(req *http.Request) (*http.Response, error) {
			sc, _ := os.ReadFile("testdata/game.json")

			return httpmock.NewStringResponse(http.StatusOK, string(sc)), nil
		},
	)

	httpmock.RegisterResponder(
		http.MethodGet,
		"https://site.api.espn.com/apis/site/v2/sports/football/nfl/summary?event=4&lang=es&region=us",
		func(req *http.Request) (*http.Response, error) {
			sc, _ := os.ReadFile("testdata/game_wrong_date.json")

			return httpmock.NewStringResponse(http.StatusOK, string(sc)), nil
		},
	)
}
