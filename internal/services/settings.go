package services

import (
	"context"

	"video-downloader-bot/internal/models"
)

type SettingsStorage interface {
	GetUser(ctx context.Context, userID int64) (*models.User, error)
	ToggleFastMode(ctx context.Context, userID int64) (*models.User, error)
	SetSettingsLanguage(ctx context.Context, userID int64, lang string) error
	CycleCaptionMode(ctx context.Context, userID int64) (*models.User, error)
}

type SettingsSvc struct {
	storage SettingsStorage
}

func NewSettingsService(ss SettingsStorage) *SettingsSvc {
	return &SettingsSvc{storage: ss}
}

func (svc *SettingsSvc) GetOrCreate(ctx context.Context, userID int64) (*models.User, error) {
	return svc.storage.GetUser(ctx, userID)
}

func (svc *SettingsSvc) ToggleFastMode(ctx context.Context, userID int64) (*models.User, error) {
	return svc.storage.ToggleFastMode(ctx, userID)
}

func (svc *SettingsSvc) SetLanguage(ctx context.Context, userID int64, lang string) error {
	return svc.storage.SetSettingsLanguage(ctx, userID, lang)
}

func (svc *SettingsSvc) CycleCaptionMode(ctx context.Context, userID int64) (*models.User, error) {
	return svc.storage.CycleCaptionMode(ctx, userID)
}
