package telegram

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/logger"
	"video-downloader-bot/internal/models"
	"video-downloader-bot/internal/telegram/helpers/errors"
	"video-downloader-bot/internal/telegram/helpers/inlines"
	"video-downloader-bot/internal/telegram/helpers/keyboards"
	"video-downloader-bot/internal/telegram/helpers/messages"
	"video-downloader-bot/internal/telegram/helpers/reactions"
	"video-downloader-bot/internal/telegram/helpers/storage"
	"video-downloader-bot/internal/telegram/helpers/videos"
	req "video-downloader-bot/internal/telegram/middlewares/requests"
	s "video-downloader-bot/internal/telegram/strings"
	"video-downloader-bot/internal/utils/links"
	"video-downloader-bot/pkg/ffmpeg"
	"video-downloader-bot/pkg/ytdlp"
)

func (b *Bot) LinkHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	reqCtx := req.FromExtContext(ctx)
	logger.DebugWithID(reqCtx, "New message \"%s\" from user %s...", ctx.EffectiveMessage.Text, storage.GetUserString(ctx.EffectiveMessage.From))

	go func() { storage.EnsureUser(ctx, b.users, new(models.UsageChat)) }()

	lang := s.GetUserLanguageCode(ctx)
	settings, err := b.settings.GetOrCreate(reqCtx, ctx.EffectiveMessage.From.Id)
	if err != nil {
		logger.ErrorWithID(reqCtx, "Failed to get settings for user %s: %v", storage.GetUserString(ctx.EffectiveUser), err)
	} else {
		lang = settings.Language
	}

	if !b.rateLimiter.Allow(ctx.EffectiveMessage.From.Id) {
		logger.WarnWithID(reqCtx, "Rate limited %s", storage.GetUserString(ctx.EffectiveMessage.From))
		messages.Reply(bot, ctx, s.Lang(lang).RateLimited)
		return nil
	}

	link, ok := links.HasLink(ctx.EffectiveMessage.Text)
	if !ok {
		logger.DebugWithID(reqCtx, "Not a link, skipping")
		messages.Reply(bot, ctx, s.Lang(lang).NoLink)
		return nil
	}

	if supported := links.IsSupportedLink(link); !supported {
		logger.DebugWithID(reqCtx, "Not a link, skipping")
		messages.Reply(bot, ctx, s.Lang(lang).UnsupportedLink)
		return nil
	}

	if err = reactions.React(bot, ctx, reactions.ReactionEyes); err != nil {
		logger.WarnWithID(reqCtx, "Failed to set reaction: %v", err)
	}

	logger.DebugWithID(reqCtx, "Found link \"%s\", processing...", link)

	if statusMsg := messages.Reply(bot, ctx, s.Lang(lang).Downloading); statusMsg != nil {
		ctx.Data["status_message_id"] = statusMsg.MessageId
		ctx.Data["status_chat_id"] = ctx.EffectiveMessage.Chat.Id
	}

	result, err := videos.DownloadAndConvert(reqCtx, link)
	if err != nil {
		return errors.HandleError(bot, ctx, lang, s.Lang(lang).FailedToProcessLink, err)
	}
	defer func() { _ = os.Remove(filepath.Join(viper.GetString(config.TelegramBotVideoDownloadFolder), filepath.Base(result))) }()
	defer func() { _ = os.Remove(filepath.Join(viper.GetString(config.TelegramBotVideoConvertedFolder), filepath.Base(result))) }()

	var statusMsgID int64
	if statusMsgID, ok = ctx.Data["status_message_id"].(int64); ok && statusMsgID != 0 {
		chatID, _ := ctx.Data["status_chat_id"].(int64)
		if _, err = bot.DeleteMessage(chatID, statusMsgID, nil); err != nil {
			logger.WarnWithID(reqCtx, "Failed to delete status message: %v", err)
		}
	}

	file, err := os.Open(result)
	if err != nil {
		return errors.HandleError(bot, ctx, lang, s.Lang(lang).FailedToProcessLink, err)
	}
	defer func() { _ = file.Close() }()

	dims, err := ffmpeg.Probe(reqCtx, result)
	if err != nil {
		logger.WarnWithID(reqCtx, "ffprobe failed, sending without dimensions: %v", err)
	}

	captionMode := models.CaptionFullDetails
	if settings != nil {
		captionMode = settings.SettingsCaptionMode
	}

	var caption, parseMode string
	switch captionMode {
	case models.CaptionVideoOnly:
		// no caption
	case models.CaptionVideoAndLink:
		caption = links.DetrackLink(link)
	default:
		var metadata *ytdlp.Metadata
		metadata, err = videos.FetchMetadata(reqCtx, link)
		if err != nil {
			logger.DebugWithID(reqCtx, "Failed to fetch metadata, falling back to link caption: %v", err)
			caption = links.DetrackLink(link)
		} else {
			caption = links.GetVideoCaption(metadata, link)
			parseMode = "HTML"
		}
	}

	logger.DebugWithID(reqCtx, "Sending video to %s...", storage.GetUserString(ctx.EffectiveMessage.From))

	sendOpts := &gotgbot.SendVideoOpts{Caption: caption, ParseMode: parseMode, SupportsStreaming: true}
	if dims != nil {
		sendOpts.Width = dims.Width
		sendOpts.Height = dims.Height
	}

	if _, err = bot.SendVideo(ctx.EffectiveMessage.Chat.Id, gotgbot.InputFileByReader(filepath.Base(result), file), sendOpts); err != nil {
		return errors.HandleError(bot, ctx, lang, s.Lang(lang).FailedToProcessLink, err)
	}

	logger.InfoWithID(reqCtx, "Video (%s) sent to %s.", link, storage.GetUserString(ctx.EffectiveMessage.From))

	if err = reactions.ReactWithTimer(bot, ctx, reactions.ReactionOK, 4); err != nil {
		logger.WarnWithID(reqCtx, "Failed to set reaction: %v", err)
	}

	return nil
}

func (b *Bot) EnteredInlineLinkHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	reqCtx := req.FromExtContext(ctx)

	if ctx.InlineQuery.Query == "" {
		logger.DebugWithID(reqCtx, "User %s opened inline menu", storage.GetUserString(&ctx.InlineQuery.From))
		return nil
	}
	logger.DebugWithID(reqCtx, "New inline query \"%s\" from %s (typing...)", ctx.InlineQuery.Query, storage.GetUserString(&ctx.InlineQuery.From))

	go func() { storage.EnsureUser(ctx, b.users, nil) }()

	if !b.rateLimiter.Allow(ctx.InlineQuery.From.Id) {
		logger.WarnWithID(reqCtx, "Rate limited %s", storage.GetUserString(&ctx.InlineQuery.From))
		return nil
	}

	if len(ctx.InlineQuery.Query) < 8 { // All supported domains are 8 chars minimum (e.g. shortest is "youtu.be/")
		return nil
	}

	var link = ctx.InlineQuery.Query
	if !links.IsOnlyLink(link) {
		if links.ContainsLink(link) {
			if links.IsSupportedLink(link) {
				link = links.ExtractLink(link)
			} else {
				logger.DebugWithID(reqCtx, "Not a link, not answering")
				return nil
			}
		} else {
			logger.DebugWithID(reqCtx, "Not a link, not answering")
			return nil
		}
	}

	logger.DebugWithID(reqCtx, "Answering inline query \"%s\" from %s...", ctx.InlineQuery.Query, storage.GetUserString(&ctx.InlineQuery.From))

	lang := s.GetUserLanguageCode(ctx)
	cleanLink := links.DetrackLink(link)
	var title, description, thumbnail string
	var fastMode bool
	settings, err := b.settings.GetOrCreate(reqCtx, ctx.InlineQuery.From.Id)
	if err != nil {
		logger.ErrorWithID(reqCtx, "Failed to get settings for user %s: %v", storage.GetUserString(ctx.EffectiveUser), err)
	} else {
		fastMode = settings.SettingsFastMode
	}

	if fastMode {
		title = s.Lang(lang).InlineFastModeTitle
		description = s.Lang(lang).InlineFastModeDescription
	} else {
		var metadata *ytdlp.Metadata
		metadata, err = videos.FetchMetadata(reqCtx, link)
		if err != nil {
			return err
		}
		thumbnail = metadata.Thumbnail
		title = metadata.Title
		description = metadata.Description
	}

	_, err = ctx.InlineQuery.Answer(bot, inlines.GetInlineResult(lang, s.Lang(lang).InlineResultPlaceholder, thumbnail, title, description, cleanLink), inlines.GetDefaultOpts())
	if err != nil {
		return errors.HandleError(bot, ctx, lang, s.Lang(lang).FailedToProcessLink, err)
	}

	logger.DebugWithID(reqCtx, "Sent article for \"%s\" to %s.", ctx.InlineQuery.Query, storage.GetUserString(&ctx.InlineQuery.From))

	return nil
}

func (b *Bot) SentInlineLinkHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.ChosenInlineResult == nil {
		return nil
	}

	link := links.ExtractLink(ctx.ChosenInlineResult.Query)
	placeholderMessageID := ctx.ChosenInlineResult.InlineMessageId
	if link == "" || placeholderMessageID == "" {
		return nil
	}

	reqCtx := req.FromExtContext(ctx)
	ctx.Data["inline_message_id"] = placeholderMessageID //
	ctx.Data["inline_link"] = links.DetrackLink(link)    //
	logger.DebugWithID(reqCtx, "Article for \"%s\" was sent to %s as a placeholder, processing...", ctx.Update.ChosenInlineResult.Query, storage.GetUserString(&ctx.Update.ChosenInlineResult.From))

	go func() { storage.EnsureUser(ctx, b.users, new(models.UsageInline)) }()

	lang := s.GetUserLanguageCode(ctx)
	settings, err := b.settings.GetOrCreate(reqCtx, ctx.Update.ChosenInlineResult.From.Id)
	if err != nil {
		logger.ErrorWithID(reqCtx, "Failed to get settings for user %s: %v", storage.GetUserString(ctx.EffectiveUser), err)
	} else {
		lang = settings.Language
	}

	result, err := videos.DownloadAndConvert(reqCtx, link)
	if err != nil {
		return errors.HandleError(bot, ctx, lang, s.Lang(lang).FailedToProcessLink, err)
	}
	defer func() { _ = os.Remove(filepath.Join(viper.GetString(config.TelegramBotVideoDownloadFolder), filepath.Base(result))) }()
	defer func() { _ = os.Remove(filepath.Join(viper.GetString(config.TelegramBotVideoConvertedFolder), filepath.Base(result))) }()

	file, err := os.Open(result)
	if err != nil {
		return errors.HandleError(bot, ctx, lang, s.Lang(lang).FailedToProcessLink, err)
	}
	defer func() { _ = file.Close() }()

	dims, err := ffmpeg.Probe(reqCtx, result)
	if err != nil {
		logger.WarnWithID(reqCtx, "ffprobe failed, sending without dimensions: %v", err)
	}

	logger.DebugWithID(reqCtx, "Sending video to dump channel...")

	sendVideoOpts := &gotgbot.SendVideoOpts{Caption: links.DetrackLink(link), SupportsStreaming: true}
	if dims != nil {
		sendVideoOpts.Width = dims.Width
		sendVideoOpts.Height = dims.Height
	}

	msg, err := bot.SendVideo(viper.GetInt64(config.TelegramBotVideoDumpChatID), gotgbot.InputFileByReader(filepath.Base(result), file), sendVideoOpts)
	if err != nil {
		return errors.HandleError(bot, ctx, lang, s.Lang(lang).FailedToProcessLink, err)
	}

	logger.DebugWithID(reqCtx, "Sent video to dump channel.")

	//noinspection GoErrorStringFormat
	if msg.Video == nil {
		return errors.HandleError(bot, ctx, lang, s.Lang(lang).FailedToProcessLink, fmt.Errorf("Telegram returned message without videos"))
	}

	captionMode := models.CaptionFullDetails
	if settings != nil {
		captionMode = settings.SettingsCaptionMode
	}

	var caption string
	switch captionMode {
	case models.CaptionVideoOnly:
		// no caption
	case models.CaptionVideoAndLink:
		caption = links.DetrackLink(link)
	default:
		var metadata *ytdlp.Metadata
		metadata, err = videos.FetchMetadata(reqCtx, link)
		if err != nil {
			logger.WarnWithID(reqCtx, "metadata fetch failed, falling back to link caption: %v", err)
			caption = links.DetrackLink(link)
		} else {
			caption = links.GetVideoCaption(metadata, link)
		}
	}

	logger.DebugWithID(reqCtx, "Attaching video to placeholder message...")

	if _, _, err = bot.EditMessageMedia(
		gotgbot.InputMediaVideo{Media: gotgbot.InputFileByID(msg.Video.FileId), Caption: caption, ParseMode: "HTML", SupportsStreaming: false},
		&gotgbot.EditMessageMediaOpts{InlineMessageId: placeholderMessageID, ReplyMarkup: keyboards.GetInlineResultButton(lang, links.DetrackLink(link))},
	); err != nil {
		return errors.HandleError(bot, ctx, lang, s.Lang(lang).FailedToProcessLink, err)
	}

	logger.InfoWithID(reqCtx, "Video (%s) sent to %s via inline mode.", link, storage.GetUserString(&ctx.Update.ChosenInlineResult.From))

	return nil
}
