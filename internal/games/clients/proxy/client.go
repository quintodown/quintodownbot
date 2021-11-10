package proxy_client

import (
	"github.com/quintodown/quintodownbot/internal/games"
)

type ProxyClient struct {
	espnClient games.GameInfoClient
}

type ClientOption func(*ProxyClient)

func WithESPNClient(espnClient games.GameInfoClient) ClientOption {
	return func(pc *ProxyClient) {
		pc.espnClient = espnClient
	}
}

func NewProxyClient(opts ...ClientOption) games.GameInfoClient {
	c := &ProxyClient{}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (p *ProxyClient) GetGames(c games.Competition) ([]games.Game, error) {
	switch c {
	default:
		return p.espnClient.GetGames(c)
	}
}

func (p *ProxyClient) GetGameInformation(c games.Competition, id string) (games.Game, error) {
	switch c {
	default:
		return p.espnClient.GetGameInformation(c, id)
	}
}
