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
