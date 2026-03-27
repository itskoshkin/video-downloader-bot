package services

import (
	"context"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"video-downloader-bot/internal/models"
)

type UserStorage interface {
	SaveUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, userID int64) (*models.User, error)
	UpdateUser(ctx context.Context, userID int64, user *models.User) error
	IncrementUsage(ctx context.Context, userID int64, usageType models.UsageType) error
	DeleteUser(ctx context.Context, userID int64) error
}

type UserServiceImpl struct {
	userStorage UserStorage
}

func NewUserService(us UserStorage) *UserServiceImpl {
	return &UserServiceImpl{userStorage: us}
}

func (svc *UserServiceImpl) SaveUser(ctx context.Context, tgUser *gotgbot.User) error {
	return svc.userStorage.SaveUser(ctx, &models.User{
		TelegramID: tgUser.Id,
		Username:   tgUser.Username,
		FirstName:  tgUser.FirstName,
		LastName:   tgUser.LastName,
		Language:   tgUser.LanguageCode,
		IsPremium:  tgUser.IsPremium,
		IsBot:      tgUser.IsBot,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		LastSeenAt: time.Now(),
	})
}

func (svc *UserServiceImpl) GetUser(ctx context.Context, tgUser *gotgbot.User) (*models.User, error) {
	return svc.userStorage.GetUser(ctx, tgUser.Id)
}

func (svc *UserServiceImpl) UpdateUser(ctx context.Context, tgUser *gotgbot.User) error {
	return svc.userStorage.UpdateUser(ctx, tgUser.Id, &models.User{
		TelegramID: tgUser.Id,
		Username:   tgUser.Username,
		FirstName:  tgUser.FirstName,
		LastName:   tgUser.LastName,
		Language:   tgUser.LanguageCode,
		IsPremium:  tgUser.IsPremium,
		IsBot:      tgUser.IsBot,
		UpdatedAt:  time.Now(),
		LastSeenAt: time.Now(),
	})
}

func (svc *UserServiceImpl) IncrementUsage(ctx context.Context, userID int64, usageType models.UsageType) error {
	return svc.userStorage.IncrementUsage(ctx, userID, usageType)
}

func (svc *UserServiceImpl) DeleteUser(ctx context.Context, tgUser *gotgbot.User) error {
	return svc.userStorage.DeleteUser(ctx, tgUser.Id)
}
