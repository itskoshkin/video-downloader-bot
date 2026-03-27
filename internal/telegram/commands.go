package telegram

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"video-downloader-bot/internal/logger"
	"video-downloader-bot/internal/telegram/helpers/errors"
	"video-downloader-bot/internal/telegram/helpers/keyboards"
	"video-downloader-bot/internal/telegram/helpers/storage"
	"video-downloader-bot/internal/telegram/middlewares/requests"
	s "video-downloader-bot/internal/telegram/strings"
	"video-downloader-bot/internal/utils/text"
)

func (b *Bot) Start(bot *gotgbot.Bot, ctx *ext.Context) error {
	reqCtx := req.FromExtContext(ctx)
	logger.DebugWithID(req.FromExtContext(ctx), "New command %s from user %s", text.Blue(ctx.EffectiveMessage.Text), storage.GetUserString(ctx.EffectiveMessage.From))

	lang := s.GetUserLanguageCode(ctx)
	message := s.LangCtx(ctx).Welcome
	if storage.EnsureUser(ctx, b.users, nil) {
		if err := b.settings.SetLanguage(reqCtx, ctx.EffectiveUser.Id, lang); err != nil {
			logger.ErrorWithID(reqCtx, "Failed to set language code \"%v\" for user %s: %v", lang, storage.GetUserString(ctx.EffectiveUser), err)
		}
	} else {
		if settings, err := b.settings.GetOrCreate(reqCtx, ctx.EffectiveUser.Id); err != nil {
			logger.ErrorWithID(reqCtx, "Failed to get settings for user %s: %v", storage.GetUserString(ctx.EffectiveUser), err)
		} else {
			lang = settings.Language
			message = s.Lang(settings.Lang()).WelcomeBack
		}
	}

	if _, err := ctx.EffectiveMessage.Reply(bot, message, nil); err != nil {
		return errors.HandleError(bot, ctx, lang, s.LangCtx(ctx).ReplyError, err)
	}

	return nil
}

func (b *Bot) Help(bot *gotgbot.Bot, ctx *ext.Context) error {
	reqCtx := req.FromExtContext(ctx)
	logger.DebugWithID(req.FromExtContext(ctx), "New command %s from user %s", text.Blue(ctx.EffectiveMessage.Text), storage.GetUserString(ctx.EffectiveMessage.From))

	lang := s.GetUserLanguageCode(ctx)
	if settings, err := b.settings.GetOrCreate(reqCtx, ctx.EffectiveUser.Id); err != nil {
		logger.ErrorWithID(reqCtx, "Failed to get settings for user %s: %v", storage.GetUserString(ctx.EffectiveUser), err)
	} else {
		lang = settings.Language
	}

	if _, err := ctx.EffectiveMessage.Reply(bot, s.Lang(lang).Help, nil); err != nil {
		return errors.HandleError(bot, ctx, lang, s.Lang(lang).ReplyError, err)
	}

	return nil
}

func (b *Bot) Settings(bot *gotgbot.Bot, ctx *ext.Context) error {
	reqCtx := req.FromExtContext(ctx)
	logger.DebugWithID(req.FromExtContext(ctx), "New command %s from user %s", text.Blue(ctx.EffectiveMessage.Text), storage.GetUserString(ctx.EffectiveMessage.From))

	storage.EnsureUser(ctx, b.users, nil)
	settings, err := b.settings.GetOrCreate(reqCtx, ctx.EffectiveUser.Id)
	if err != nil {
		return errors.HandleError(bot, ctx, s.GetUserLanguageCode(ctx), s.LangCtx(ctx).SettingsError, err)
	}

	if _, err = ctx.EffectiveMessage.Reply(bot, s.Lang(settings.Lang()).Settings, &gotgbot.SendMessageOpts{ReplyMarkup: keyboards.GetSettingsKeyboard(settings)}); err != nil {
		return errors.HandleError(bot, ctx, settings.Lang(), s.Lang(settings.Lang()).SettingsError, err)
	}

	return nil
}

func (b *Bot) Farewell(bot *gotgbot.Bot, ctx *ext.Context) error {
	reqCtx := req.FromExtContext(ctx)
	logger.DebugWithID(req.FromExtContext(ctx), "New command %s from user %s", text.Blue(ctx.EffectiveMessage.Text), storage.GetUserString(ctx.EffectiveMessage.From))

	lang := s.GetUserLanguageCode(ctx)
	if settings, err := b.settings.GetOrCreate(reqCtx, ctx.EffectiveUser.Id); err != nil {
		logger.ErrorWithID(reqCtx, "Failed to get settings for user %s: %v", storage.GetUserString(ctx.EffectiveUser), err)
	} else {
		lang = settings.Language
	}

	if _, err := ctx.EffectiveMessage.Reply(bot, s.Lang(lang).Farewell, &gotgbot.SendMessageOpts{ReplyMarkup: keyboards.GetForgetConfirmationKeyboard(lang)}); err != nil {
		return errors.HandleError(bot, ctx, lang, s.Lang(lang).SettingsError, err)
	}

	return nil
}
