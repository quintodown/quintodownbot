package proxy_client_test

import (
	"github.com/quintodown/quintodownbot/internal/games"
	proxy_client "github.com/quintodown/quintodownbot/internal/games/clients/proxy"
	mgames "github.com/quintodown/quintodownbot/mocks/games"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestProxyClient_GetGames(t *testing.T) {
	espn := new(mgames.GameInfoClient)

	espn.On("GetGames", games.NFL).Once().Return([]games.Game{
		{
			Id: "12345",
		},
	}, nil)

	gms, err := proxy_client.NewProxyClient(proxy_client.WithESPNClient(espn)).
		GetGames(games.NFL)

	require.NoError(t, err)
	require.Len(t, gms, 1)
	espn.AssertExpectations(t)
}

func TestProxyClient_GetGameInformation(t *testing.T) {
	espn := new(mgames.GameInfoClient)

	start := time.Now().UTC()
	espn.On("GetGameInformation", games.NFL, "12345").Once().Return(games.Game{
		Id:    "12345",
		Start: start,
	}, nil)

	gm, err := proxy_client.NewProxyClient(proxy_client.WithESPNClient(espn)).
		GetGameInformation(games.NFL, "12345")

	require.NoError(t, err)
	require.Equal(t, games.Game{
		Id:    "12345",
		Start: start,
	}, gm)
	espn.AssertExpectations(t)
}
