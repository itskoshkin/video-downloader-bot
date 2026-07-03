package app

import (
	"context"
	"time"

	"gorm.io/gorm"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/logger"
	"video-downloader-bot/internal/providers"
	"video-downloader-bot/internal/services"
	"video-downloader-bot/internal/storage"
	"video-downloader-bot/internal/telegram"
	"video-downloader-bot/pkg/ffmpeg"
	"video-downloader-bot/pkg/postgres"
	"video-downloader-bot/pkg/ytdlp"
)

type App struct {
	db  *gorm.DB
	bot *telegram.Bot
}

func Load() *App {
	config.LoadConfig()
	logger.SetupLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var err error

	// External tools
	err = ytdlp.CheckIfInstalled(ctx)
	if err != nil {
		logger.Fatal(err)
	}
	err = ffmpeg.CheckIfInstalled(ctx)
	if err != nil {
		logger.Fatal(err)
	}
	err = config.EnsureWorkingDirs()
	if err != nil {
		logger.Fatal(err)
	}

	// Databases/clients
	db, err := postgres.NewInstance(config.PostgresConfig())
	if err != nil {
		logger.Fatal(err)
	}

	// Storages
	userStore := storage.NewUserStorage(db)

	// Services
	userSvc := services.NewUserService(userStore)
	setsSvc := services.NewSettingsService(userStore)

	// Providers
	manager := providers.NewManager(
		providers.NewYtDlp(),
		providers.NewHikerAPI(),
		providers.NewInstagrapi(),
		providers.NewPreview(),
	)

	// Telegram
	bot := telegram.NewBot(userSvc, setsSvc, manager)

	return &App{bot: bot, db: db}
}

func (a *App) Run() {
	a.bot.RegisterHandlers()
	a.bot.Run()
	a.bot.Idle()
	a.shutdown()
}

func (a *App) shutdown() {
	logger.Info("Shutting down...")

	a.bot.Stop()
	if sqlDB, err := a.db.DB(); err == nil {
		_ = sqlDB.Close()
	}

	logger.Info("Shutdown complete.")
	logger.Close()
}
