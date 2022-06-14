//go:build wireinject
// +build wireinject

package app

import (
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/quintodown/quintodownbot/internal/clock"
	"github.com/quintodown/quintodownbot/internal/games"
	"github.com/quintodown/quintodownbot/internal/games/clients/espn"
	proxyclient "github.com/quintodown/quintodownbot/internal/games/clients/proxy"
	"github.com/quintodown/quintodownbot/internal/handlers"
	handlersgames "github.com/quintodown/quintodownbot/internal/handlers/games"

	"github.com/quintodown/quintodownbot/internal/telegram"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	hse "github.com/quintodown/quintodownbot/internal/handlers/error"
	hstl "github.com/quintodown/quintodownbot/internal/handlers/telegram"
	hstw "github.com/quintodown/quintodownbot/internal/handlers/twitter"
	"github.com/quintodown/quintodownbot/internal/pubsub"
	"github.com/sirupsen/logrus"

	"github.com/dghubble/oauth1"
	"github.com/google/wire"
	"github.com/quintodown/quintodownbot/internal/bot"
	"github.com/quintodown/quintodownbot/internal/config"
	"github.com/quintodown/quintodownbot/internal/twitter"

	gt "github.com/javiyt/go-twitter/twitter"
	tb "gopkg.in/telebot.v3"
)

const (
	updateGamesInformationTicker = time.Minute
	updateGamesListTicker        = 6 * time.Hour
	bufferEventTime              = 5 * time.Second
)

type customHandlerGenerator func() []handlers.EventHandler

var (
	queueInstance *gochannel.GoChannel
	twitterClient = wire.NewSet(
		provideTwitterHttpClient,
		provideTwitterClient,
		wire.Bind(new(bot.TwitterClient), new(*twitter.Client)),
	)
	queue        = wire.NewSet(provideGoChannelQueue, wire.Bind(new(pubsub.Queue), new(*gochannel.GoChannel)))
	telegramDeps = wire.NewSet(provideConfiguration, provideTBot, queue)
	twitterDeps  = wire.NewSet(provideConfiguration, twitterClient, queue)
	errorDeps    = wire.NewSet(provideConfiguration, queue, provideLogger)
	gamesDeps    = wire.NewSet(
		wire.NewSet(clock.NewUTCClock, wire.Bind(new(clock.Clock), new(clock.UTCClock))),
		queue,
		provideGameInfoClient,
		provideGameHandler,
	)
)

func ProvideApp() (*App, func(), error) {
	panic(wire.Build(
		provideBotProvider,
		initializeCustomHandlers,
		provideHandlers,
		wire.NewSet(queue, provideHandlerManager),
		NewApp,
	))
}

func provideBotProvider() botProvider {
	return provideBot
}

func provideBot() (bot.AppBot, error) {
	panic(wire.Build(
		provideConfiguration,
		provideTBot,
		twitterClient,
		queue,
		provideBotOptions,
		bot.NewBot,
	))
}

func provideConfiguration() (config.AppConfig, error) {
	panic(wire.Build(config.NewAppConfig))
}

func provideTBot() (bot.TelegramBot, error) {
	panic(wire.Build(provideConfiguration, provideTBotSettings, tb.NewBot, telegram.NewBot))
}

func provideTBotSettings(cfg config.AppConfig) tb.Settings {
	return tb.Settings{
		Token:  cfg.BotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}
}

func provideTwitterClient(*http.Client) *twitter.Client {
	wire.Build(gt.NewClient, twitter.NewTwitterClient)

	return &twitter.Client{}
}

func provideTwitterHttpClient(cfg config.AppConfig) *http.Client {
	return oauth1.NewConfig(cfg.TwitterAPIKey, cfg.TwitterAPISecret).
		Client(oauth1.NoContext, oauth1.NewToken(cfg.TwitterAccessToken, cfg.TwitterAccessSecret))
}

func provideGoChannelQueue() *gochannel.GoChannel {
	if queueInstance == nil {
		queueInstance = gochannel.NewGoChannel(
			gochannel.Config{},
			watermill.NewStdLogger(true, true),
		)
	}
	return queueInstance
}

func provideBotOptions(b bot.TelegramBot, cfg config.AppConfig, tc bot.TwitterClient, gq pubsub.Queue) []bot.Option {
	return []bot.Option{
		bot.WithTelegramBot(b),
		bot.WithConfig(cfg),
		bot.WithTwitterClient(tc),
		bot.WithQueue(gq),
	}
}

func provideLogger(cfg config.AppConfig) (*logrus.Logger, func()) {
	var (
		file *os.File
		err  error
	)

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	logger.SetReportCaller(true)

	lvl := logrus.DebugLevel
	if cfg.IsProd() {
		lvl = logrus.ErrorLevel

		if cfg.LogFile != "" {
			file, err = os.OpenFile(cfg.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o755)
			if err != nil {
				logger.Fatal(err)
			}
			logger.SetOutput(file)
		}
	}

	logger.SetLevel(lvl)

	return logger, func() {
		logger.Exit(0)
		_ = file.Close()
	}
}

func provideTelegramOptions(cfg config.AppConfig, tb bot.TelegramBot, pq pubsub.Queue) []hstl.Option {
	return []hstl.Option{
		hstl.WithAppConfig(cfg),
		hstl.WithTelegramBot(tb),
		hstl.WithQueue(pq),
	}
}

func provideTelegramHandler() (*hstl.Telegram, error) {
	panic(wire.Build(telegramDeps, provideTelegramOptions, hstl.NewTelegram))
}

func provideTwitterOptions(tc bot.TwitterClient, pq pubsub.Queue) []hstw.Option {
	return []hstw.Option{
		hstw.WithTwitterClient(tc),
		hstw.WithQueue(pq),
	}
}

func provideTwitterHandler() (*hstw.Twitter, error) {
	panic(wire.Build(twitterDeps, provideTwitterOptions, hstw.NewTwitter))
}

func provideErrorHandler() (*hse.ErrorHandler, func(), error) {
	panic(wire.Build(errorDeps, hse.NewErrorHandler))
}

func provideHandlers(customHandlers customHandlerGenerator) ([]handlers.EventHandler, func(), error) {
	telegramHandler, err := provideTelegramHandler()
	if err != nil {
		return nil, nil, err
	}
	twitterHandler, err := provideTwitterHandler()
	if err != nil {
		return nil, nil, err
	}
	errorHandler, cleanup, err := provideErrorHandler()
	if err != nil {
		return nil, nil, err
	}

	return append(customHandlers(),
		telegramHandler,
		twitterHandler,
		errorHandler,
	), cleanup, nil
}

func provideHandlerManager(q pubsub.Queue, h []handlers.EventHandler) *handlers.Manager {
	return handlers.NewHandlersManager(q, h...)
}

func initializeCustomHandlers() customHandlerGenerator {
	return func() []handlers.EventHandler {
		gamesHandler, _ := provideGames()

		return []handlers.EventHandler{gamesHandler}
	}
}

func provideGameOptions(gh games.Handler, q pubsub.Queue) []handlersgames.Option {
	return []handlersgames.Option{
		handlersgames.WithGameHandler(gh),
		handlersgames.WithConfig(handlersgames.Config{
			UpdateGamesInformationTicker: updateGamesInformationTicker,
			UpdateGamesListTicker:        updateGamesListTicker,
			BufferEventTime:              bufferEventTime,
		}),
		handlersgames.WithQueue(q),
	}
}

func provideGames() (*handlersgames.Games, error) {
	panic(wire.Build(gamesDeps, provideGameOptions, handlersgames.NewGames))
}

func provideGameHandler(gc games.GameInfoClient, q pubsub.Queue, clk clock.Clock) games.Handler {
	return games.NewGameHandler(gc, true, q, clk)
}

func provideGameInfoClient(clk clock.Clock) games.GameInfoClient {
	return proxyclient.NewProxyClient(proxyclient.WithESPNClient(espn.NewESPNClient(provideHTTClient(), clk)))
}

func provideHTTClient() *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	httpClient := retryClient.HTTPClient
	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return httpClient
}
