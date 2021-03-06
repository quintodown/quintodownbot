package app_test

import (
	"context"
	"os"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	pubsub2 "github.com/quintodown/quintodownbot/internal/pubsub"
	"github.com/quintodown/quintodownbot/mocks/pubsub"

	"github.com/quintodown/quintodownbot/internal/app"
	"github.com/quintodown/quintodownbot/internal/bot"
	"github.com/quintodown/quintodownbot/internal/handlers"
	"github.com/stretchr/testify/require"

	mockBot "github.com/quintodown/quintodownbot/mocks/bot"
)

type startAppError struct{}

func (m startAppError) Error() string {
	return "could not start"
}

type botInstanceError struct{}

func (m botInstanceError) Error() string {
	return "bot instance not ready"
}

func TestInitializeConfiguration(t *testing.T) {
	envFile := []byte("BOT_TOKEN=asdfg")
	envTestFile := []byte("BOT_TOKEN=qwert")

	t.Run("it should load configuration from environment file when not in test env", func(t *testing.T) {
		e := app.InitializeConfiguration(false, envFile, envTestFile)

		require.NoError(t, e)
		require.Equal(t, "asdfg", os.Getenv("BOT_TOKEN"))
		_ = os.Unsetenv("BOT_TOKEN")
	})

	t.Run("it should load test configuration from environment file when in test env", func(t *testing.T) {
		e := app.InitializeConfiguration(true, envFile, envTestFile)

		require.NoError(t, e)
		require.Equal(t, "qwert", os.Getenv("BOT_TOKEN"))
		_ = os.Unsetenv("BOT_TOKEN")
	})

	t.Run("it should fail when not a valid env file", func(t *testing.T) {
		e := app.InitializeConfiguration(true, []byte("BOT_TOKEN"), envTestFile)

		require.EqualError(t, e, "error loading env file: line `BOT_TOKEN` doesn't match format")
	})

	t.Run("it should fail when not a valid env.test file", func(t *testing.T) {
		e := app.InitializeConfiguration(true, envFile, []byte("BOT_TOKEN"))

		require.EqualError(t, e, "error loading env.test file: line `BOT_TOKEN` doesn't match format")
		_ = os.Unsetenv("BOT_TOKEN")
	})
}

func TestStart(t *testing.T) {
	t.Run("it should fail when getting bot instance", func(t *testing.T) {
		mbp := func() (bot.AppBot, error) {
			return nil, botInstanceError{}
		}

		a := app.NewApp(mbp, handlers.NewHandlersManager(nil))
		e := a.Start(context.Background())

		require.EqualError(t, e, "error getting bot instance: bot instance not ready")
	})

	t.Run("it should fail when starting bot instance", func(t *testing.T) {
		mb := new(mockBot.AppBot)
		mbp := func() (bot.AppBot, error) {
			return mb, nil
		}
		mb.On("Start", context.Background()).Once().Return(startAppError{})

		a := app.NewApp(mbp, handlers.NewHandlersManager(nil))
		e := a.Start(context.Background())

		require.EqualError(t, e, "error starting bot: could not start")
		mb.AssertExpectations(t)
	})

	t.Run("it should start bot instance", func(t *testing.T) {
		q := new(pubsub.Queue)
		mb := new(mockBot.AppBot)
		mbp := func() (bot.AppBot, error) {
			return mb, nil
		}
		mb.On("Start", context.Background()).Once().Return(nil)
		q.On("Subscribe", context.Background(), pubsub2.CommandTopic.String()).
			Return(func(context.Context, string) <-chan *message.Message {
				return make(chan *message.Message)
			}, nil)

		a := app.NewApp(mbp, handlers.NewHandlersManager(q))
		e := a.Start(context.Background())

		require.NoError(t, e)
		mb.AssertExpectations(t)
	})
}

func TestRun(t *testing.T) {
	q := new(pubsub.Queue)
	mb := new(mockBot.AppBot)
	mbp := func() (bot.AppBot, error) {
		return mb, nil
	}

	mb.On("Start", context.Background()).Once().Return(nil)
	mb.On("Run").Once()
	q.On("Subscribe", context.Background(), pubsub2.CommandTopic.String()).
		Return(func(context.Context, string) <-chan *message.Message {
			return make(chan *message.Message)
		}, nil)

	a := app.NewApp(mbp, handlers.NewHandlersManager(q))
	_ = a.Start(context.Background())
	a.Run()

	mb.AssertExpectations(t)
}

func TestStop(t *testing.T) {
	q := new(pubsub.Queue)
	mb := new(mockBot.AppBot)
	mbp := func() (bot.AppBot, error) {
		return mb, nil
	}

	mb.On("Start", context.Background()).Once().Return(nil)
	mb.On("Stop").Once()
	q.On("Subscribe", context.Background(), pubsub2.CommandTopic.String()).
		Return(func(context.Context, string) <-chan *message.Message {
			return make(chan *message.Message)
		}, nil)

	a := app.NewApp(mbp, handlers.NewHandlersManager(q))
	e := a.Start(context.Background())
	a.Stop()

	require.NoError(t, e)
	mb.AssertExpectations(t)
}
