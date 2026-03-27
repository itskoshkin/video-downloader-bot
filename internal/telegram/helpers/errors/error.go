package errors

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/logger"
	"video-downloader-bot/internal/telegram/helpers/reactions"
	"video-downloader-bot/internal/telegram/middlewares/requests"
	s "video-downloader-bot/internal/telegram/strings"
	"video-downloader-bot/internal/utils/errs"
)

func HandlerErrorHandler() func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
	return func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
		if isToolDebugEnabled(err) {
			logger.ErrorWithID(req.FromExtContext(ctx), "gotgbot: handler: unhandled error: %v", err)
		} else {
			logger.ErrorWithID(req.FromExtContext(ctx), "gotgbot: handler: unhandled error: %v", errs.ShortError(err))
		}
		return ext.DispatcherActionNoop
	}
}

func DispatcherErrorHandler() func(err error) {
	return func(err error) {
		logger.Error("gotgbot: dispatcher: %v", err)
	}
}

func DispatcherPanicHandler() func(b *gotgbot.Bot, ctx *ext.Context, r any) {
	return func(b *gotgbot.Bot, ctx *ext.Context, r any) {
		logger.ErrorWithID(req.FromExtContext(ctx), "gotgbot: dispatcher: panic recovered: %v\n%s", r, debug.Stack())
	}
}

func UpdaterErrorHandler() func(err error) {
	return func(err error) {
		logger.Error("gotgbot: updater: %v", err)
	}
}

func HandleError(bot *gotgbot.Bot, ctx *ext.Context, languageCode string, message string, err error) error {
	detail := errs.ShortStderr(err)
	if detail != "" {
		logger.ErrorWithID(req.FromExtContext(ctx), "%s", detail)
	} else {
		logger.ErrorWithID(req.FromExtContext(ctx), "%v", err)
	}

	if stderr := errs.FullStderr(err); stderr != "" && isToolDebugEnabled(err) {
		logger.ErrorWithFileID(req.FromExtContext(ctx), "full stderr:\n%s", stderr)
	}

	errorLine := "❌ " + message
	if detail != "" {
		errorLine += "\n" + detail
	}

	requestID, _ := ctx.Data["request_id"].(string)

	var replyMarkup *gotgbot.InlineKeyboardMarkup
	if requestID != "" {
		replyMarkup = &gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{Text: s.Lang(languageCode).RequestID + ": " + requestID, CopyText: &gotgbot.CopyTextButton{Text: requestID}},
				},
				//{
				//	{Text: "Get support", Url: "https://t.me/feed_the_cat_bot?start=" + requestID}, //TODO: Enable
				//},
			},
		}
	}

	if statusMsgID, ok := ctx.Data["status_message_id"].(int64); ok && statusMsgID != 0 {
		// Edit the "⏳ Downloading..." status message that was sent before processing
		chatID, _ := ctx.Data["status_chat_id"].(int64)
		_ = reactions.ReactWithTimer(bot, ctx, reactions.ReactionBrokenHeart, 10)
		editOpts := &gotgbot.EditMessageTextOpts{ChatId: chatID, MessageId: statusMsgID}
		if replyMarkup != nil {
			editOpts.ReplyMarkup = *replyMarkup
		}
		if _, _, editErr := bot.EditMessageText(errorLine, editOpts); editErr != nil {
			return fmt.Errorf("original errors: %v; edit status message errors: %w", err, editErr)
		}
	} else if ctx.EffectiveMessage != nil {
		_ = reactions.ReactWithTimer(bot, ctx, reactions.ReactionBrokenHeart, 10)
		if _, replyErr := ctx.EffectiveMessage.Reply(bot, errorLine, &gotgbot.SendMessageOpts{ReplyMarkup: replyMarkup}); replyErr != nil {
			return fmt.Errorf("original errors: %v; reply errors: %w", err, replyErr)
		}
	} else if inlineMessageID, ok := ctx.Data["inline_message_id"].(string); ok && inlineMessageID != "" {
		inlineLink, _ := ctx.Data["inline_link"].(string)
		inlineText := errorLine
		if inlineLink != "" {
			inlineText += "\n\n" + inlineLink
		}
		editOpts := &gotgbot.EditMessageTextOpts{InlineMessageId: inlineMessageID}
		if replyMarkup != nil {
			editOpts.ReplyMarkup = *replyMarkup
		}
		if _, _, editErr := bot.EditMessageText(inlineText, editOpts); editErr != nil {
			return fmt.Errorf("original errors: %v; edit message errors: %w", err, editErr)
		}
	}

	return err
}

func isToolDebugEnabled(err error) bool {
	binary := errs.ExecBinary(err)
	switch {
	case strings.Contains(binary, "ffmpeg") || strings.Contains(binary, "ffprobe"):
		return viper.GetBool(config.FfmpegDebug)
	case strings.Contains(binary, "yt-dlp"):
		return viper.GetBool(config.YtDlpDebug)
	default:
		return true
	}
}
