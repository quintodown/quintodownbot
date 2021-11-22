package handlersgames

import (
	"context"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/mailru/easyjson"
	"github.com/quintodown/quintodownbot/internal/games"
	"github.com/quintodown/quintodownbot/internal/handlers"
	"github.com/quintodown/quintodownbot/internal/pubsub"
)

type Games struct {
	gh games.Handler
	c  Config
	q  pubsub.Queue
}

type Config struct {
	UpdateGamesInformationTicker time.Duration
}

type Option func(g *Games)

func WithGameHandler(gh games.Handler) Option {
	return func(g *Games) {
		g.gh = gh
	}
}

func WithConfig(c Config) Option {
	return func(g *Games) {
		g.c = c
	}
}

func WithQueue(q pubsub.Queue) Option {
	return func(g *Games) {
		g.q = q
	}
}

func NewGames(options ...Option) *Games {
	g := &Games{}

	for _, o := range options {
		o(g)
	}

	return g
}

func (g *Games) ExecuteHandlers(ctx context.Context) {
	g.updateGamesInformation(ctx)
}

func (g *Games) updateGamesInformation(ctx context.Context) {
	messages, err := g.q.Subscribe(ctx, pubsub.GamesTopic.String())
	if err != nil {
		handlers.SendError(g.q, err)

		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.Tick(g.c.UpdateGamesInformationTicker):
				g.gh.UpdateGamesInformation(true)
			}
		}
	}()

	go g.sendGameUpdate(messages)
}

func (g *Games) sendGameUpdate(messages <-chan *message.Message) {
	for msg := range messages {
		var m pubsub.GameEvent

		if err := easyjson.Unmarshal(msg.Payload, &m); err != nil {
			handlers.SendError(g.q, err)
			msg.Ack()

			continue
		}

		gameText := ""

		switch m.LastGameChange {
		case games.Started.String():
			gameText = g.getStartedGameMessage(m)
		case games.Finished.String():
			gameText = g.getFinishedGameMessage(m)
		default:
			msg.Ack()

			continue
		}

		mb, _ := easyjson.Marshal(pubsub.TextEvent{Text: gameText})

		if err := g.q.Publish(pubsub.TextTopic.String(), message.NewMessage(watermill.NewUUID(), mb)); err != nil {
			handlers.SendError(g.q, err)
			msg.Nack()

			continue
		}

		msg.Ack()
	}
}

func (g *Games) getStartedGameMessage(m pubsub.GameEvent) string {
	return fmt.Sprintf(
		"El partido entre %s (%s) vs %s (%s) ha iniciado. Se juega en %s (%s, %s)",
		m.AwayTeam.Name,
		m.AwayTeam.Record,
		m.HomeTeam.Name,
		m.HomeTeam.Record,
		m.Venue.FullName,
		m.Venue.City,
		m.Venue.State,
	)
}

func (g *Games) getFinishedGameMessage(m pubsub.GameEvent) string {
	return fmt.Sprintf(
		"El partido entre %s (%s) vs %s (%s) ha finalizado con el resultado de %v - %v",
		m.AwayTeam.Name,
		m.AwayTeam.Record,
		m.HomeTeam.Name,
		m.HomeTeam.Record,
		m.AwayTeam.Score,
		m.HomeTeam.Score,
	)
}
