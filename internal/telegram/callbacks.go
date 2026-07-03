package telegram

import (
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"video-downloader-bot/internal/logger"
	"video-downloader-bot/internal/models"
	"video-downloader-bot/internal/providers"
	"video-downloader-bot/internal/telegram/helpers/errors"
	"video-downloader-bot/internal/telegram/helpers/keyboards"
	"video-downloader-bot/internal/telegram/helpers/storage"
	req "video-downloader-bot/internal/telegram/middlewares/requests"
	s "video-downloader-bot/internal/telegram/strings"
)

func (b *Bot) SettingsCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	reqCtx := req.FromExtContext(ctx)
	callback := ctx.CallbackQuery

	storage.EnsureUser(ctx, b.users, nil)
	settings, err := b.settings.GetOrCreate(reqCtx, ctx.EffectiveUser.Id)
	if err != nil {
		return errors.HandleError(bot, ctx, s.GetUserLanguageCode(ctx), s.LangCtx(ctx).SettingsError, err)
	}

	editSettingsMessage := func(bot *gotgbot.Bot, cb *gotgbot.CallbackQuery, settings *models.User) error {
		_, _, err = cb.Message.EditText(bot, s.Lang(settings.Lang()).Settings, &gotgbot.EditMessageTextOpts{ReplyMarkup: *keyboards.GetSettingsKeyboard(settings)})
		return err
	}

	switch {

	case callback.Data == "settings:fast_mode":
		settings, err = b.settings.ToggleFastMode(reqCtx, callback.From.Id)
		if err != nil {
			logger.ErrorWithID(reqCtx, "failed to toggle fast mode: %v", err)
			_, _ = callback.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: s.Lang(settings.Lang()).SettingsError})
			return nil
		}

		if err = editSettingsMessage(bot, callback, settings); err != nil {
			logger.ErrorWithID(reqCtx, "failed to edit settings message: %v", err)
		}

		_, _ = callback.Answer(bot, nil)

	case callback.Data == "settings:caption_mode":
		settings, err = b.settings.CycleCaptionMode(reqCtx, callback.From.Id)
		if err != nil {
			logger.ErrorWithID(reqCtx, "failed to cycle caption mode: %v", err)
			_, _ = callback.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: s.Lang(settings.Lang()).SettingsError})
			return nil
		}

		if err = editSettingsMessage(bot, callback, settings); err != nil {
			logger.ErrorWithID(reqCtx, "failed to edit settings message: %v", err)
		}

		_, _ = callback.Answer(bot, nil)

	case callback.Data == "settings:language":
		_, _, err = callback.Message.EditReplyMarkup(bot, &gotgbot.EditMessageReplyMarkupOpts{
			ReplyMarkup: *keyboards.GetLanguageKeyboard(settings),
		})
		if err != nil {
			logger.ErrorWithID(reqCtx, "failed to edit language keyboard: %v", err)
		}

		_, _ = callback.Answer(bot, nil)

	case strings.HasPrefix(callback.Data, "settings:lang:"):
		selectedLang := strings.TrimPrefix(callback.Data, "settings:lang:")

		if err = b.settings.SetLanguage(reqCtx, callback.From.Id, selectedLang); err != nil {
			logger.ErrorWithID(reqCtx, "failed to set language: %v", err)
			_, _ = callback.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: s.Lang(settings.Lang()).SettingsError})
			return nil
		}

		settings, err = b.settings.GetOrCreate(reqCtx, callback.From.Id)
		if err != nil {
			logger.ErrorWithID(reqCtx, "failed to get settings: %v", err)
			return nil
		}

		if err = editSettingsMessage(bot, callback, settings); err != nil {
			logger.ErrorWithID(reqCtx, "failed to edit settings message: %v", err)
		}

		_, _ = callback.Answer(bot, nil)

	case callback.Data == "settings:back":
		settings, err = b.settings.GetOrCreate(reqCtx, callback.From.Id)
		if err != nil {
			logger.ErrorWithID(reqCtx, "failed to get settings: %v", err)
			return nil
		}

		if err = editSettingsMessage(bot, callback, settings); err != nil {
			logger.ErrorWithID(reqCtx, "failed to edit settings message: %v", err)
		}

		_, _ = callback.Answer(bot, nil)

	}

	return nil
}

func (b *Bot) FarewellCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	reqCtx := req.FromExtContext(ctx)
	callback := ctx.CallbackQuery

	storage.EnsureUser(ctx, b.users, nil)
	settings, err := b.settings.GetOrCreate(reqCtx, ctx.EffectiveUser.Id)
	if err != nil {
		return errors.HandleError(bot, ctx, s.GetUserLanguageCode(ctx), s.LangCtx(ctx).SettingsError, err)
	}

	switch {

	case callback.Data == "farewell:confirm":
		if _, _, err = bot.EditMessageText(s.Lang(settings.Lang()).FarewellInProgress, &gotgbot.EditMessageTextOpts{ChatId: ctx.EffectiveMessage.Chat.Id, MessageId: ctx.EffectiveMessage.MessageId}); err != nil {
			logger.ErrorWithID(reqCtx, "failed to edit farewell message: %v", err)
		}

		if err = b.users.DeleteUser(reqCtx, ctx.EffectiveUser); err != nil {
			logger.ErrorWithID(reqCtx, "failed to delete user: %v", err)
		}

		if _, _, err = bot.EditMessageText(s.Lang(settings.Lang()).FarewellDone, &gotgbot.EditMessageTextOpts{ChatId: ctx.EffectiveMessage.Chat.Id, MessageId: ctx.EffectiveMessage.MessageId}); err != nil {
			logger.ErrorWithID(reqCtx, "failed to edit farewell message: %v", err)
		}

		if _, _, err = bot.EditMessageText(s.Lang(settings.Lang()).FarewellGoodbye, &gotgbot.EditMessageTextOpts{ChatId: ctx.EffectiveMessage.Chat.Id, MessageId: ctx.EffectiveMessage.MessageId}); err != nil {
			logger.ErrorWithID(reqCtx, "failed to edit farewell message: %v", err)
		}

		_, _ = callback.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: s.Lang(settings.Lang()).FarewellComplete})

	case callback.Data == "farewell:cancel":
		if _, err = bot.DeleteMessage(ctx.EffectiveMessage.Chat.Id, ctx.EffectiveMessage.MessageId, nil); err != nil {
			logger.ErrorWithID(reqCtx, "failed to delete farewell message: %v", err)
		}

		if _, err = bot.DeleteMessage(ctx.EffectiveMessage.Chat.Id, ctx.EffectiveMessage.MessageId-1, nil); err != nil {
			logger.ErrorWithID(reqCtx, "failed to delete farewell command: %v", err)
		}

		_, _ = callback.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: s.Lang(settings.Lang()).FarewellCancel})

	}

	return nil
}

// PreviewCallback handles the "preview not working?" button: it cycles a URL-rewrite preview to the
// next embed-fix domain. Callback data is "preview:<idx>:<detracked-link>".
func (b *Bot) PreviewCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	reqCtx := req.FromExtContext(ctx)
	callback := ctx.CallbackQuery

	idxStr, link, found := strings.Cut(strings.TrimPrefix(callback.Data, "preview:"), ":")
	idx, err := strconv.Atoi(idxStr)
	if !found || err != nil {
		_, _ = callback.Answer(bot, nil)
		return nil
	}

	rewritten, ok := providers.PreviewURL(link, idx)
	if !ok {
		_, _ = callback.Answer(bot, nil)
		return nil
	}

	lang := s.GetUserLanguageCode(ctx)
	if settings, sErr := b.settings.GetOrCreate(reqCtx, callback.From.Id); sErr == nil {
		lang = settings.Language
	}

	markup := keyboards.GetPreviewKeyboard(lang, "https://"+link, providers.PreviewCycleData(link, idx+1))
	editOpts := &gotgbot.EditMessageTextOpts{ReplyMarkup: markup}
	if callback.InlineMessageId != "" {
		editOpts.InlineMessageId = callback.InlineMessageId
		_, _, err = bot.EditMessageText(rewritten, editOpts)
	} else if callback.Message != nil {
		_, _, err = callback.Message.EditText(bot, rewritten, editOpts)
	}
	if err != nil {
		logger.WarnWithID(reqCtx, "failed to cycle preview: %v", err)
	}

	_, _ = callback.Answer(bot, nil)
	return nil
}
