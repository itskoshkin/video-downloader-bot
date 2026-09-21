package inlines

import (
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/telegram/helpers/keyboards"
	s "video-downloader-bot/internal/telegram/strings"
	"video-downloader-bot/internal/utils/links"
)

// GetInlineResult builds the inline article and the placeholder message it sends. Title/description
// are the picker UI; the message itself is headline + processing notice + link. The headline is the
// same head as the finished caption and is omitted entirely when we have no metadata (fast mode).
func GetInlineResult(languageCode, thumbnail, title, description, headline, link string) []gotgbot.InlineQueryResult {
	if description == "" {
		description = s.GetLocalizedString(languageCode).InlineSendingHint
	}

	parts := make([]string, 0, 3)
	if headline != "" {
		parts = append(parts, headline)
	}
	parts = append(parts, s.GetLocalizedString(languageCode).InlineProcessingNotice, links.DetrackLink(link))

	content := gotgbot.InputTextMessageContent{MessageText: strings.Join(parts, "\n\n")}
	if headline != "" {
		content.ParseMode = "HTML" // the headline carries <b> from the caption builder
	}

	return []gotgbot.InlineQueryResult{
		gotgbot.InlineQueryResultArticle{
			Id:                  "download",
			ThumbnailUrl:        thumbnail,
			Title:               title,
			Description:         description,
			InputMessageContent: content,
			ReplyMarkup:         keyboards.GetInlinePlaceholderButton(languageCode),
		},
	}
}

func GetDefaultOpts() *gotgbot.AnswerInlineQueryOpts {
	return &gotgbot.AnswerInlineQueryOpts{
		IsPersonal: false,
		CacheTime:  new(viper.GetInt64(config.TelegramBotInlineCacheTime)),
	}
}
