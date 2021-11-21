package handlersgames_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/mailru/easyjson"
	games2 "github.com/quintodown/quintodownbot/internal/games"
	"github.com/quintodown/quintodownbot/internal/handlers"
	handlersgames "github.com/quintodown/quintodownbot/internal/handlers/games"
	"github.com/quintodown/quintodownbot/internal/pubsub"
	"github.com/quintodown/quintodownbot/mocks/games"
	mps "github.com/quintodown/quintodownbot/mocks/pubsub"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const ticker = 10 * time.Millisecond

func TestGames_ExecuteHandlersFails(t *testing.T) {
	q := new(mps.Queue)
	q.On("Subscribe", context.Background(), pubsub.GamesTopic.String()).Once().
		Return(nil, errors.New("getting channel error"))
	q.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
		return string(m.Payload) == "{\"error\":\"getting channel error\"}"
	})).Once().
		Return(nil)

	g := handlersgames.NewGames(
		handlersgames.WithGameHandler(new(games.Handler)),
		handlersgames.WithConfig(handlersgames.Config{UpdateGamesInformationTicker: ticker}),
		handlersgames.WithQueue(q),
	)

	g.ExecuteHandlers(context.Background())

	q.AssertExpectations(t)
}

func TestGames_ExecuteHandlersGamesFails(t *testing.T) {
	t.Run("it does nothing cause no games playing", func(t *testing.T) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		g, _, q, gh := initGameHandlerAndMocks(ctx)

		called := make(chan interface{})

		gh.On("UpdateGamesInformation", true).Run(func(mock.Arguments) { called <- true })

		g.ExecuteHandlers(ctx)
		<-called
		cancelFunc()

		gh.AssertExpectations(t)
		q.AssertExpectations(t)
		close(called)
	})

	t.Run("it fails when message couldn't be parsed", func(t *testing.T) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		g, c, q, gh := initGameHandlerAndMocks(ctx)

		called := make(chan interface{})

		gh.On("UpdateGamesInformation", true).Run(func(mock.Arguments) {
			sendMessageToChannel(t, c, []byte("{["), true)
			called <- true
		})
		q.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"error\":\"parse error: EOF reached while skipping array/object or token "+
				"near offset 2 of ''\"}"
		})).Once().Return(nil)

		g.ExecuteHandlers(ctx)
		<-called
		cancelFunc()

		gh.AssertExpectations(t)
		q.AssertExpectations(t)
		close(called)
	})

	t.Run("it does nothing when game status not handled", func(t *testing.T) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		g, c, q, gh := initGameHandlerAndMocks(ctx)

		called := make(chan interface{})
		defer close(called)
		gh.On("UpdateGamesInformation", true).Run(func(mock.Arguments) {
			b, _ := easyjson.Marshal(pubsub.GameEvent{LastGameChange: games2.HomeScore.String()})
			sendMessageToChannel(t, c, b, true)
			called <- true
		})

		g.ExecuteHandlers(ctx)
		<-called
		cancelFunc()

		gh.AssertExpectations(t)
		q.AssertExpectations(t)
	})
}

func TestGames_ExecuteHandlersGames(t *testing.T) {
	t.Run("it sends game message", func(t *testing.T) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		g, c, q, gh := initGameHandlerAndMocks(ctx)

		called := make(chan interface{})

		gh.On("UpdateGamesInformation", true).Run(func(mock.Arguments) {
			b, _ := easyjson.Marshal(pubsub.GameEvent{
				LastGameChange: games2.GameFinished.String(),
				HomeTeam:       pubsub.TeamScore{Name: "Home Team", Score: 1, Record: "1-2"},
				AwayTeam:       pubsub.TeamScore{Name: "Away Team", Score: 2, Record: "2-1"},
			})
			sendMessageToChannel(t, c, b, true)

			called <- true
		})
		q.On("Publish", pubsub.TextTopic.String(), mock.MatchedBy(func(message *message.Message) bool {
			return string(message.Payload) == "{\"text\":\"El partido entre Away Team (2-1) vs Home Team (1-2) ha "+
				"finalizado con el resultado de 2 - 1\"}"
		})).Once().Return(nil)

		g.ExecuteHandlers(ctx)
		<-called
		cancelFunc()

		gh.AssertExpectations(t)
		q.AssertExpectations(t)
		close(called)
	})
}

func initGameHandlerAndMocks(ctx context.Context) (
	handlers.EventHandler,
	chan *message.Message,
	*mps.Queue,
	*games.Handler,
) {
	q := new(mps.Queue)
	gh := new(games.Handler)

	gamesChannel := make(chan *message.Message)

	q.On("Subscribe", ctx, pubsub.GamesTopic.String()).Once().
		Return(func(context.Context, string) <-chan *message.Message {
			return gamesChannel
		}, nil)

	g := handlersgames.NewGames(
		handlersgames.WithGameHandler(gh),
		handlersgames.WithConfig(handlersgames.Config{UpdateGamesInformationTicker: ticker}),
		handlersgames.WithQueue(q),
	)

	return g, gamesChannel, q, gh
}

func sendMessageToChannel(t *testing.T, channel chan *message.Message, eventMsg []byte, acked bool) {
	newMessage := message.NewMessage(watermill.NewUUID(), eventMsg)
	channel <- newMessage

	require.Eventually(t, func() bool {
		if acked {
			<-newMessage.Acked()
		} else {
			<-newMessage.Nacked()
		}

		return true
	}, time.Second, time.Millisecond)
}
