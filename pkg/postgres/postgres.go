package postgres

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"video-downloader-bot/internal/models"
	"video-downloader-bot/internal/utils/text"
)

type Config struct {
	Host         string
	Port         string
	User         string
	Password     string
	Database     string
	SSLMode      string
	LogLevel     string
	MaxIdleConns int
	MaxOpenConns int
}

func NewInstance(cfg Config) (*gorm.DB, error) {
	fmt.Print("Connecting to Postgres...")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		fmt.Println()
		return nil, fmt.Errorf("GORM: failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println()
		return nil, fmt.Errorf("GORM: failed to get underlying sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	if err = sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		fmt.Println()
		return nil, fmt.Errorf("GORM: failed to ping database: %v", err)
	}

	if err = db.AutoMigrate(&models.User{}); err != nil {
		_ = sqlDB.Close()
		fmt.Println()
		return nil, fmt.Errorf("GORM: failed to auto-migrate: %v", err)
	}

	fmt.Println(text.Green(" Done."))

	return db, nil
}
