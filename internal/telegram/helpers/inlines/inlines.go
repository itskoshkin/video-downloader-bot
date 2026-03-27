package inlines

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/telegram/helpers/keyboards"
	s "video-downloader-bot/internal/telegram/strings"
	"video-downloader-bot/internal/utils/links"
)

func GetInlineResult(languageCode, template, thumbnail, title, description, link string) []gotgbot.InlineQueryResult {
	if description == "" {
		description = s.GetLocalizedString(languageCode).InlineSendingHint
	}

	return []gotgbot.InlineQueryResult{
		gotgbot.InlineQueryResultArticle{
			Id:                  "download",
			ThumbnailUrl:        thumbnail,
			Title:               title,
			Description:         description,
			InputMessageContent: gotgbot.InputTextMessageContent{MessageText: fmt.Sprintf(template, title, description, links.DetrackLink(link))},
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
