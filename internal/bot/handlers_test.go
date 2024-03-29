package bot_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/quintodown/quintodownbot/internal/pubsub"
	mq "github.com/quintodown/quintodownbot/mocks/pubsub"

	"github.com/quintodown/quintodownbot/internal/bot"
	"github.com/quintodown/quintodownbot/internal/config"
	mb "github.com/quintodown/quintodownbot/mocks/bot"
	"github.com/stretchr/testify/mock"

	tb "gopkg.in/telebot.v3"
)

const (
	adminID          = 12345
	broadcastChannel = int64(987654)
	imagePayload     = "{\"caption\":\"testing\"," +
		"\"fileId\":\"blablabla\"," +
		"\"fileUrl\":\"https://myimage.com/test.jpg\"," +
		"\"fileSize\":1234," +
		"\"fileContent\":\"iVBORw0KGgoAAAANSUhEUgAAAAQAAAAECAIAAAAmkwkpAAAAQklEQVR4nGJWTd9ZaWdyOfW69Y8z" +
		"DF5sfALun5c7SL+8ysQUqp7euSxThUtU5v9FJg2PoueTrrw5Vyt36AYgAAD//yOnFnjB+cHEAAAAAElFTkSuQmCC\"}"
)

type downloadImageError struct{}

func (m downloadImageError) Error() string {
	return "error downloading image"
}

func TestHandlerStartAndHelpCommand(t *testing.T) {
	commands := []struct {
		command  string
		expected string
	}{
		{
			command:  "/start",
			expected: "Thanks for using the bot! You can type /help command to know what can I do",
		},
		{
			command:  "/help",
			expected: "/help - Show help\n/start - Start a conversation with the bot\n",
		},
	}
	for i := range commands {
		i := i
		handler, mockedBot, _ := generateHandlerAndMockedBot(t, commands[i].command, config.AppConfig{})

		t.Run("it should do nothing when not in private conversation", func(t *testing.T) {
			_ = handler(bot.TelegramMessage{
				IsPrivate: false,
			})

			mockedBot.AssertExpectations(t)
			mockedBot.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
		})

		t.Run("it should message when in private conversation", func(t *testing.T) {
			m := bot.TelegramMessage{IsPrivate: true, SenderID: "1234"}
			mockedBot.On("Send", m.SenderID, commands[i].expected).Once().Return(nil, nil)

			_ = handler(m)

			mockedBot.AssertExpectations(t)
		})
	}

	t.Run("it should fail when help requested by non numeric user id", func(t *testing.T) {
		handler, _, _ := generateHandlerAndMockedBot(t, "/help", config.AppConfig{})
		m := bot.TelegramMessage{IsPrivate: true, SenderID: "asdf"}

		_ = handler(m)
	})

	t.Run("it should send admin commands when user admin", func(t *testing.T) {
		handler, mockedBot, _ := generateHandlerAndMockedBot(t, "/help", config.AppConfig{Admins: []int{1234}})
		m := bot.TelegramMessage{IsPrivate: true, SenderID: "1234"}
		expected := "/help - Show help\n/start - Start a conversation with the bot\n/stop - Stop notifications" +
			" for all handlers or specific handler\n"
		mockedBot.On("Send", m.SenderID, expected).Once().Return(nil, nil)

		_ = handler(m)

		mockedBot.AssertExpectations(t)
	})
}

func TestHandlersFilters(t *testing.T) {
	commands := []string{tb.OnPhoto, tb.OnText}
	for i := range commands {
		i := i
		adminID := 12345
		broadcastChannel := int64(987654)

		mockedQueue := new(mq.Queue)
		handler, mockedBot, _ := generateHandlerAndMockedBot(t, commands[i], config.AppConfig{
			Admins:           []int{adminID},
			BroadcastChannel: broadcastChannel,
		})

		testCases := []struct {
			name string
			m    bot.TelegramMessage
		}{
			{
				name: "it should do nothing when not in private conversation",
				m: bot.TelegramMessage{
					IsPrivate: false,
					SenderID:  "1234",
				},
			},
			{
				name: "it should do nothing when in private conversation but not admin",
				m: bot.TelegramMessage{
					IsPrivate: true,
					SenderID:  "54321",
				},
			},
			{
				name: "it should fail when in private conversation but sender can't be converted to int",
				m: bot.TelegramMessage{
					IsPrivate: true,
					SenderID:  "asdfg",
				},
			},
		}

		for i := range testCases {
			i := i
			t.Run(testCases[i].name, func(t *testing.T) {
				_ = handler(testCases[i].m)

				mockedBot.AssertExpectations(t)
				mockedQueue.Test(t)
				mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
			})
		}
	}
}

func TestHandlerPhoto(t *testing.T) {
	handler, mockedBot, mockedQueue := generateHandlerAndMockedBot(t, tb.OnPhoto, config.AppConfig{
		Admins:           []int{adminID},
		BroadcastChannel: broadcastChannel,
	})

	successPhoto := bot.TelegramMessage{
		IsPrivate: true,
		SenderID:  strconv.Itoa(adminID),
		Photo: bot.TelegramPhoto{
			Caption:  "testing",
			FileID:   "blablabla",
			FileURL:  "https://myimage.com/test.jpg",
			FileSize: 1234,
		},
	}

	t.Run("it should do nothing when caption no present", func(t *testing.T) {
		_ = handler(bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  strconv.Itoa(adminID),
			Photo:     bot.TelegramPhoto{Caption: ""},
		})

		mockedBot.AssertExpectations(t)
		mockedQueue.Test(t)
		mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	})

	t.Run("it should do nothing when error getting image", func(t *testing.T) {
		mockedBot.On("GetFile", successPhoto.Photo.FileID).Once().
			Return(nil, downloadImageError{})

		_ = handler(successPhoto)

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	})

	t.Run("it should send photo when caption is present and image could be downloaded", func(t *testing.T) {
		file, _ := os.Open("testdata/test.png")
		defer func() { _ = file.Close() }()

		mockedBot.On("GetFile", successPhoto.Photo.FileID).Once().Return(file, nil)
		mockedQueue.On(
			"Publish",
			pubsub.PhotoTopic.String(),
			mock.MatchedBy(func(message *message.Message) bool {
				return string(message.Payload) == imagePayload
			}),
		).Once().Return(nil)

		_ = handler(successPhoto)

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertExpectations(t)
	})
}

func TestHandlerText(t *testing.T) {
	handler, mockedBot, mockedQueue := generateHandlerAndMockedBot(t, tb.OnText, config.AppConfig{
		Admins:           []int{adminID},
		BroadcastChannel: broadcastChannel,
	})

	t.Run("it should do nothing when text no present", func(t *testing.T) {
		_ = handler(bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  strconv.Itoa(adminID),
			Text:      "",
		})

		mockedBot.AssertExpectations(t)
		mockedQueue.Test(t)
		mockedQueue.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("it should send text when present", func(t *testing.T) {
		m := bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  strconv.Itoa(adminID),
			Text:      "testing",
		}
		mockedQueue.On(
			"Publish",
			pubsub.TextTopic.String(),
			mock.MatchedBy(func(message *message.Message) bool {
				return string(message.Payload) == "{\"text\":\"testing\"}"
			}),
		).Once().Return(nil)

		_ = handler(m)

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertExpectations(t)
	})
}

func TestHandleStopNotifications(t *testing.T) {
	handler, mockedBot, mockedQueue := generateHandlerAndMockedBot(t, "/stop", config.AppConfig{
		Admins:           []int{adminID},
		BroadcastChannel: broadcastChannel,
	})

	t.Run("it should send an stop command event", func(t *testing.T) {
		mockedQueue.On("Publish", pubsub.CommandTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"command\":0,\"handler\":\"\"}"
		})).Once().Return(nil)
		_ = handler(bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  strconv.Itoa(adminID),
			Text:      "/stop",
		})

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should send an stop command event to particular handle", func(t *testing.T) {
		mockedQueue.On("Publish", pubsub.CommandTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"command\":0,\"handler\":\"telegram\"}"
		})).Once().Return(nil)
		_ = handler(bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  strconv.Itoa(adminID),
			Text:      "/stop telegram",
			Payload:   "telegram",
		})

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertExpectations(t)
	})
}

func generateHandlerAndMockedBot(
	t *testing.T,
	toHandle string,
	cfg config.AppConfig,
) (bot.TelegramHandler, *mb.TelegramBot, *mq.Queue) {
	allHandlers := []string{"/start", "/help", "/stop", tb.OnPhoto, tb.OnText}

	var (
		handler bot.TelegramHandler
		ok      bool
	)

	mockedQueue := new(mq.Queue)

	mockedBot := new(mb.TelegramBot)
	mockedBot.On("SetCommands", mock.Anything).Once().Return(nil)

	for _, v := range allHandlers {
		if v == toHandle {
			mockedBot.On("Handle", toHandle, mock.Anything).
				Once().
				Return(nil, nil).
				Run(func(args mock.Arguments) {
					handler, ok = args.Get(1).(bot.TelegramHandler)
					if !ok {
						t.Fatal("given handler is not valid")
					}
				})
		} else {
			mockedBot.On("Handle", v, mock.Anything).Once().Return(nil, nil)
		}
	}

	_ = bot.NewBot(
		bot.WithTelegramBot(mockedBot),
		bot.WithConfig(cfg),
		bot.WithQueue(mockedQueue),
	).Start(nil)

	return handler, mockedBot, mockedQueue
}
