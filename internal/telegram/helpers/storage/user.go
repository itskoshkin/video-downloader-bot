package storage

import (
	"context"
	"errors"
	"fmt"
	"video-downloader-bot/internal/utils/text"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"gorm.io/gorm"

	"video-downloader-bot/internal/logger"
	"video-downloader-bot/internal/models"
	"video-downloader-bot/internal/telegram/middlewares/requests"
)

type UserService interface {
	SaveUser(ctx context.Context, tgUser *gotgbot.User) error
	GetUser(ctx context.Context, tgUser *gotgbot.User) (*models.User, error)
	UpdateUser(ctx context.Context, tgUser *gotgbot.User) error
	IncrementUsage(ctx context.Context, userID int64, usageType models.UsageType) error
}

func EnsureUser(ctx *ext.Context, svc UserService, usage *models.UsageType) (isNewUser bool) {
	reqCtx := req.FromExtContext(ctx)

	existing, err := svc.GetUser(reqCtx, ctx.EffectiveUser)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			isNewUser = true
			logger.DebugWithID(reqCtx, "User %s not found, creating...", GetUserString(ctx.EffectiveUser))
			if saveErr := svc.SaveUser(reqCtx, ctx.EffectiveUser); saveErr != nil {
				logger.ErrorWithID(reqCtx, "Failed to save user %s: %v", GetUserString(ctx.EffectiveUser), saveErr)
			}
		} else {
			isNewUser = true //
			logger.ErrorWithID(reqCtx, "Failed to get user %s: %v", GetUserString(ctx.EffectiveUser), err)
		}
	}

	if existing != nil {
		if existing.Username != ctx.EffectiveUser.Username || existing.FirstName != ctx.EffectiveUser.FirstName || existing.LastName != ctx.EffectiveUser.LastName || existing.IsPremium != ctx.EffectiveUser.IsPremium {
			if updateErr := svc.UpdateUser(reqCtx, ctx.EffectiveUser); updateErr != nil {
				logger.ErrorWithID(reqCtx, "Failed to update user %s: %v", GetUserString(ctx.EffectiveUser), updateErr)
			}
		}
	}

	if usage != nil {
		if err = svc.IncrementUsage(reqCtx, ctx.EffectiveUser.Id, *usage); err != nil {
			logger.ErrorWithID(reqCtx, "Failed to increment usage for %s: %v", GetUserString(ctx.EffectiveUser), err)
		}
	}

	return isNewUser
}

func GetUserString(tgUser *gotgbot.User) string {
	var name string
	if tgUser.FirstName != "" {
		name += tgUser.FirstName
	}
	if tgUser.LastName != "" {
		name += " " + tgUser.LastName
	}

	var details string
	if tgUser.Username != "" {
		details += "@" + tgUser.Username + ", "
	}
	if tgUser.Id != 0 {
		details += fmt.Sprintf("ID %d", tgUser.Id)
	}

	return text.Blue(fmt.Sprintf("%s (%s)", name, details))
}
