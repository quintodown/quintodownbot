package handlerserror

import (
	"context"

	"github.com/mailru/easyjson"
	"github.com/quintodown/quintodownbot/internal/pubsub"
	"github.com/sirupsen/logrus"
)

type ErrorHandler struct {
	log *logrus.Logger
	q   pubsub.Queue
}

func NewErrorHandler(log *logrus.Logger, q pubsub.Queue) *ErrorHandler {
	return &ErrorHandler{log: log, q: q}
}

func (eh *ErrorHandler) ID() string {
	return "error"
}

func (eh *ErrorHandler) ExecuteHandlers(ctx context.Context) {
	messages, err := eh.q.Subscribe(ctx, pubsub.ErrorTopic.String())
	if err != nil {
		eh.log.Error(err)
	}

	go func() {
		for msg := range messages {
			var m pubsub.ErrorEvent

			if err := easyjson.Unmarshal(msg.Payload, &m); err != nil {
				eh.log.Error(err)
				msg.Ack()

				continue
			}

			eh.log.Error(m.Err)
			msg.Ack()
		}
	}()
}

func (eh *ErrorHandler) StopNotifications() {
	return
}
