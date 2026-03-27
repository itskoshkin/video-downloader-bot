package storage

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"video-downloader-bot/internal/models"
)

type UserStore struct {
	db *gorm.DB
}

func NewUserStorage(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) SaveUser(ctx context.Context, user *models.User) error {
	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to save user with ID %d: %w", user.TelegramID, err)
	}

	return nil
}

func (s *UserStore) GetUser(ctx context.Context, telegramID int64) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user with ID %d not found: %w", telegramID, err)
	}

	return &user, nil
}

func (s *UserStore) UpdateUser(ctx context.Context, telegramID int64, user *models.User) error {
	if err := s.db.WithContext(ctx).Model(&models.User{}).Where("telegram_id = ?", telegramID).Select("username", "first_name", "last_name", "language", "is_premium", "is_bot", "updated_at", "last_seen_at").Updates(user).Error; err != nil {
		return fmt.Errorf("failed to update user with ID %d: %w", telegramID, err)
	}

	return nil
}

func (s *UserStore) IncrementUsage(ctx context.Context, telegramID int64, usageType models.UsageType) error {
	updates := map[string]any{
		"total_use_count": gorm.Expr("total_use_count + 1"),
		"last_seen_at":    gorm.Expr("NOW()"),
	}

	switch usageType {
	case models.UsageChat:
		updates["chat_use_count"] = gorm.Expr("chat_use_count + 1")
	case models.UsageInline:
		updates["inline_use_count"] = gorm.Expr("inline_use_count + 1")
	}

	if err := s.db.WithContext(ctx).Model(&models.User{}).Where("telegram_id = ?", telegramID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to increment usage count for user %d: %w", telegramID, err)
	}

	return nil
}

func (s *UserStore) ToggleFastMode(ctx context.Context, telegramID int64) (*models.User, error) {
	if err := s.db.WithContext(ctx).Model(&models.User{}).Where("telegram_id = ?", telegramID).Update("settings_fast_mode", gorm.Expr("NOT settings_fast_mode")).Error; err != nil {
		return nil, fmt.Errorf("failed to toggle fast mode for user %d: %w", telegramID, err)
	}

	return s.GetUser(ctx, telegramID)
}

func (s *UserStore) SetSettingsLanguage(ctx context.Context, telegramID int64, lang string) error {
	if err := s.db.WithContext(ctx).Model(&models.User{}).Where("telegram_id = ?", telegramID).Update("settings_language", lang).Error; err != nil {
		return fmt.Errorf("failed to set language for user %d: %w", telegramID, err)
	}

	return nil
}

func (s *UserStore) CycleCaptionMode(ctx context.Context, telegramID int64) (*models.User, error) {
	user, err := s.GetUser(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	next := user.SettingsCaptionMode.Next()
	if err = s.db.WithContext(ctx).Model(&models.User{}).Where("telegram_id = ?", telegramID).Update("settings_caption_mode", next).Error; err != nil {
		return nil, fmt.Errorf("failed to cycle caption mode for user %d: %w", telegramID, err)
	}

	user.SettingsCaptionMode = next
	return user, nil
}

func (s *UserStore) DeleteUser(ctx context.Context, telegramID int64) error {
	if err := s.db.WithContext(ctx).Model(&models.User{}).Where("telegram_id = ?", telegramID).Delete(&models.User{}).Error; err != nil {
		return fmt.Errorf("failed to delete user with ID %d: %w", telegramID, err)
	}

	return nil
}
