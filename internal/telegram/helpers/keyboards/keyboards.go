package keyboards

import (
	"github.com/PaulSonOfLars/gotgbot/v2"

	"video-downloader-bot/internal/models"
	s "video-downloader-bot/internal/telegram/strings"
)

func GetInlinePlaceholderButton(languageCode string) *gotgbot.InlineKeyboardMarkup {
	return &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: s.Lang(languageCode).InlineButtonDownloading, CallbackData: "pending"}, //TODO: Use cb.Answer to show progress?
			},
		},
	}
}

func GetInlineResultButton(languageCode, link string) gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: s.Lang(languageCode).InlineButtonOpenLink, Url: link},
			},
		},
	}
}

// GetPreviewKeyboard builds the keyboard for a URL-rewrite preview: an "open link" button to the
// original post plus a "preview not working?" button that cycles to the next embed-fix domain
// (omitted when cycleData is empty, e.g. only one domain configured).
func GetPreviewKeyboard(languageCode, openURL, cycleData string) gotgbot.InlineKeyboardMarkup {
	rows := [][]gotgbot.InlineKeyboardButton{
		{{Text: s.Lang(languageCode).InlineButtonOpenLink, Url: openURL}},
	}
	if cycleData != "" {
		rows = append(rows, []gotgbot.InlineKeyboardButton{
			{Text: s.Lang(languageCode).PreviewBrokenButton, CallbackData: cycleData},
		})
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func GetSettingsKeyboard(user *models.User) *gotgbot.InlineKeyboardMarkup {
	fastMode := s.Lang(user.Lang()).SettingsFastMode + ": "
	if user.SettingsFastMode {
		fastMode += s.Lang(user.Lang()).SettingsFastModeEnabled
	} else {
		fastMode += s.Lang(user.Lang()).SettingsFastModeDisabled
	}

	caption := s.Lang(user.Lang()).SettingsCaption + ": " + user.SettingsCaptionMode.Label(user.Lang())
	language := s.Lang(user.Lang()).SettingsLanguage + ": " + user.Lang()

	return &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{{Text: fastMode, CallbackData: "settings:fast_mode"}},
			{{Text: caption, CallbackData: "settings:caption_mode"}},
			{{Text: language, CallbackData: "settings:language"}},
		},
	}
}

func GetLanguageKeyboard(user *models.User) *gotgbot.InlineKeyboardMarkup {
	return &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "🇬🇧 English", CallbackData: "settings:lang:en"},
				{Text: "🇷🇺 Русский", CallbackData: "settings:lang:ru"},
				{Text: "🇺🇦 Українська", CallbackData: "settings:lang:ua"},
			},
			{{Text: s.Lang(user.Lang()).SettingsBackButton, CallbackData: "settings:back"}},
		},
	}
}

func GetForgetConfirmationKeyboard(languageCode string) *gotgbot.InlineKeyboardMarkup {
	return &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: s.Lang(languageCode).FarewellConfirm, CallbackData: "farewell:confirm"},
			},
			{
				{Text: s.Lang(languageCode).FarewellCancel, CallbackData: "farewell:cancel"},
			},
		},
	}
}
