package handlersgames

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/mailru/easyjson"
	"github.com/quintodown/quintodownbot/internal/games"
	"github.com/quintodown/quintodownbot/internal/handlers"
	"github.com/quintodown/quintodownbot/internal/pubsub"
)

type Games struct {
	gh           games.Handler
	c            Config
	q            pubsub.Queue
	shouldNotify bool
}

type Config struct {
	UpdateGamesInformationTicker time.Duration
	UpdateGamesListTicker        time.Duration
	BufferEventTime              time.Duration
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
	g := &Games{shouldNotify: true}

	for _, o := range options {
		o(g)
	}

	return g
}

func (g *Games) ID() string {
	return "games"
}

func (g *Games) StopNotifications() {
	g.shouldNotify = false
}

func (g *Games) ExecuteHandlers(ctx context.Context) {
	g.updateGamesInformation(ctx)
	g.updateGameList(ctx)
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

	go g.sendGameUpdate(ctx, messages)
}

func (g *Games) updateGameList(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.Tick(g.c.UpdateGamesListTicker):
				g.gh.UpdateGamesList()
			}
		}
	}()
}

func (g *Games) sendGameUpdate(ctx context.Context, messages <-chan *message.Message) {
	gameText := ""
	t := time.Tick(g.c.BufferEventTime)

	for {
		if !g.shouldNotify {
			continue
		}

		select {
		case msg := <-messages:
			var m pubsub.GameEvent

			if err := easyjson.Unmarshal(msg.Payload, &m); err != nil {
				handlers.SendError(g.q, err)
				msg.Ack()

				continue
			}

			switch m.LastGameChange {
			case games.Started.String():
				gameText += "\n" + g.getStartedGameMessage(m)
			case games.Finished.String():
				gameText += "\n" + g.getFinishedGameMessage(m)
			default:
				msg.Ack()

				continue
			}

			gameText = strings.Trim(gameText, "\n")

			msg.Ack()
		case <-t:
			if gameText != "" {
				mb, _ := easyjson.Marshal(pubsub.TextEvent{Text: gameText})

				if err := g.q.Publish(pubsub.TextTopic.String(), message.NewMessage(watermill.NewUUID(), mb)); err != nil {
					handlers.SendError(g.q, err)
				}

				gameText = ""
			}
		case <-ctx.Done():
			return
		}
	}
}

func (g *Games) getStartedGameMessage(m pubsub.GameEvent) string {
	return fmt.Sprintf(
		"#%s El partido entre %s (%s) vs %s (%s) ha iniciado. Se juega en %s (%s, %s)",
		m.Competition,
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
		"#%s El partido entre %s (%s) vs %s (%s) ha finalizado con el resultado de %v - %v",
		m.Competition,
		m.AwayTeam.Name,
		m.AwayTeam.Record,
		m.HomeTeam.Name,
		m.HomeTeam.Record,
		m.AwayTeam.Score,
		m.HomeTeam.Score,
	)
}
