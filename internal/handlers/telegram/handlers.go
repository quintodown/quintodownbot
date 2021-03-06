package handlerstelegram

import (
	"context"
	"strconv"

	"github.com/mailru/easyjson"
	"github.com/quintodown/quintodownbot/internal/bot"
	"github.com/quintodown/quintodownbot/internal/config"
	"github.com/quintodown/quintodownbot/internal/handlers"
	"github.com/quintodown/quintodownbot/internal/pubsub"
)

type Telegram struct {
	bot          bot.TelegramBot
	cfg          config.AppConfig
	q            pubsub.Queue
	shouldNotify bool
}

type Option func(b *Telegram)

func WithTelegramBot(tb bot.TelegramBot) Option {
	return func(b *Telegram) {
		b.bot = tb
	}
}

func WithAppConfig(cfg config.AppConfig) Option {
	return func(b *Telegram) {
		b.cfg = cfg
	}
}

func WithQueue(q pubsub.Queue) Option {
	return func(b *Telegram) {
		b.q = q
	}
}

func NewTelegram(options ...Option) *Telegram {
	t := &Telegram{shouldNotify: true}

	for _, o := range options {
		o(t)
	}

	return t
}

func (t *Telegram) ID() string {
	return "telegram"
}

func (t *Telegram) ExecuteHandlers(ctx context.Context) {
	t.handleText(ctx)
	t.handlePhoto(ctx)
}

func (t *Telegram) StopNotifications() {
	t.shouldNotify = false
}

func (t *Telegram) handleText(ctx context.Context) {
	messages, err := t.q.Subscribe(ctx, pubsub.TextTopic.String())
	if err != nil {
		handlers.SendError(t.q, err)
	}

	go func() {
		for msg := range messages {
			if !t.shouldNotify {
				msg.Ack()

				continue
			}

			var m pubsub.TextEvent

			if err := easyjson.Unmarshal(msg.Payload, &m); err != nil {
				handlers.SendError(t.q, err)
				msg.Ack()

				continue
			}

			if err := t.bot.Send(strconv.Itoa(int(t.cfg.BroadcastChannel)), m.Text); err != nil {
				handlers.SendError(t.q, err)
			}

			msg.Ack()
		}
	}()
}

func (t *Telegram) handlePhoto(ctx context.Context) {
	messages, err := t.q.Subscribe(ctx, pubsub.PhotoTopic.String())
	if err != nil {
		handlers.SendError(t.q, err)
	}

	go func() {
		for msg := range messages {
			if !t.shouldNotify {
				msg.Ack()

				continue
			}

			var m pubsub.PhotoEvent
			if err := easyjson.Unmarshal(msg.Payload, &m); err != nil {
				handlers.SendError(t.q, err)
				msg.Ack()

				continue
			}

			if err := t.bot.Send(strconv.Itoa(int(t.cfg.BroadcastChannel)), &bot.TelegramPhoto{
				Caption:  m.Caption,
				FileID:   m.FileID,
				FileURL:  m.FileURL,
				FileSize: m.FileSize,
			}); err != nil {
				handlers.SendError(t.q, err)
			}

			msg.Ack()
		}
	}()
}
