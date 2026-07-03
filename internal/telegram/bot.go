package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/choseninlineresult"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/inlinequery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/logger"
	"video-downloader-bot/internal/models"
	"video-downloader-bot/internal/providers"
	"video-downloader-bot/internal/telegram/helpers/errors"
	"video-downloader-bot/internal/telegram/middlewares/ratelimit"
	"video-downloader-bot/internal/utils/text"
)

type UserService interface {
	SaveUser(ctx context.Context, tgUser *gotgbot.User) error
	GetUser(ctx context.Context, tgUser *gotgbot.User) (*models.User, error)
	UpdateUser(ctx context.Context, tgUser *gotgbot.User) error
	DeleteUser(ctx context.Context, tgUser *gotgbot.User) error
	IncrementUsage(ctx context.Context, userID int64, usageType models.UsageType) error
}

type SettingsService interface {
	GetOrCreate(ctx context.Context, userID int64) (*models.User, error)
	ToggleFastMode(ctx context.Context, userID int64) (*models.User, error)
	SetLanguage(ctx context.Context, userID int64, lang string) error
	CycleCaptionMode(ctx context.Context, userID int64) (*models.User, error)
}

type Bot struct {
	bot        *gotgbot.Bot
	updater    *ext.Updater
	dispatcher *ext.Dispatcher

	users       UserService
	settings    SettingsService
	rateLimiter *ratelimit.RateLimiter
	manager     *providers.Manager
}

func NewBot(users UserService, settings SettingsService, manager *providers.Manager) *Bot {
	bot, err := gotgbot.NewBot(viper.GetString(config.TelegramBotToken), nil)
	if err != nil {
		fmt.Println()
		logger.Fatalf("gotgbot: failed to create new bot: %v", err)
	}

	bridgeLogger := logger.NewBridgeLogger()
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		MaxRoutines:      ext.DefaultMaxRoutines,
		Logger:           bridgeLogger,
		Error:            errors.HandlerErrorHandler(),
		UnhandledErrFunc: errors.DispatcherErrorHandler(),
		Panic:            errors.DispatcherPanicHandler(),
	})

	updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{
		Logger:           bridgeLogger,
		UnhandledErrFunc: errors.UpdaterErrorHandler(),
	})

	rateLimiter := ratelimit.NewRateLimiter(
		viper.GetInt(config.TelegramBotRateLimitPerMinute),
		viper.GetInt(config.TelegramBotRateLimitBurst),
		viper.GetInt(config.TelegramBotRateLimitPerDay),
	)

	return &Bot{bot: bot, dispatcher: dispatcher, updater: updater, users: users, settings: settings, rateLimiter: rateLimiter, manager: manager}
}

func (b *Bot) RegisterHandlers() {
	b.registerMiddleware()

	// Commands
	b.dispatcher.AddHandler(handlers.NewCommand("start", b.Start))
	b.dispatcher.AddHandler(handlers.NewCommand("help", b.Help))
	b.dispatcher.AddHandler(handlers.NewCommand("settings", b.Settings))
	b.dispatcher.AddHandler(handlers.NewCommand("farewell", b.Farewell))

	// Callbacks
	b.dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("settings:"), b.SettingsCallback))
	b.dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("farewell:"), b.FarewellCallback))
	b.dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("preview:"), b.PreviewCallback))

	// Messages
	b.dispatcher.AddHandler(handlers.NewMessage(message.Text, b.LinkHandler))

	// Inline
	b.dispatcher.AddHandler(handlers.NewInlineQuery(inlinequery.All, b.EnteredInlineLinkHandler))
	b.dispatcher.AddHandler(handlers.NewChosenInlineResult(choseninlineresult.All, b.SentInlineLinkHandler))
}

func (b *Bot) Run() {
	fmt.Printf("Starting %s bot...", text.Blue("Telegram"))

	if err := b.updater.StartPolling(b.bot, &ext.PollingOpts{
		DropPendingUpdates: viper.GetBool(config.TelegramBotDropPendingUpdates),
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout:     viper.GetInt64(config.TelegramBotLongPollingTimeout),
			RequestOpts: &gotgbot.RequestOpts{Timeout: viper.GetDuration(config.TelegramBotHttpClientTimeout) * time.Second}},
	}); err != nil {
		fmt.Println()
		logger.Fatalf("gotgbot: failed to start polling: %v", err)
	}

	fmt.Println(text.Green("  Done."))
}

func (b *Bot) Idle() {
	logger.Info("Logged in as %s, now listening...", text.Blue("@"+b.bot.User.Username))
	b.updater.Idle()
}

func (b *Bot) Stop() {
	logger.Info("Stopping %s bot...", text.Blue("Telegram"))
	if err := b.updater.Stop(); err != nil {
		logger.Error("failed to stop updater: %v", err)
	} else {
		fmt.Println(text.Green("  Done."))
	}
}
