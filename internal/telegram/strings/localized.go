package strings

import (
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

const __ = ""

type Strings struct {
	//a, b, c, d, e, f, s string

	/* Commands */
	Welcome                     string
	WelcomeBack                 string
	Help                        string
	Settings                    string
	SettingsFastMode            string
	SettingsFastModeEnabled     string
	SettingsFastModeDisabled    string
	SettingsCaption             string
	SettingsCaptionVideoOnly    string
	SettingsCaptionVideoAndLink string
	SettingsCaptionFull         string
	SettingsLanguage            string
	SettingsBackButton          string
	Farewell                    string
	FarewellConfirm             string
	FarewellCancel              string
	FarewellInProgress          string
	FarewellDone                string
	FarewellComplete            string
	FarewellCanceled            string
	FarewellGoodbye             string

	/* Errors*/
	RateLimited   string
	SettingsError string
	RequestID     string

	/* Messages */
	ReplyError string

	/* Links */
	NoLink              string
	UnsupportedLink     string
	Downloading         string
	FailedToProcessLink string

	/* Inline */
	InlineSendingHint         string
	InlineResultPlaceholder   string
	InlineButtonDownloading   string
	InlineButtonOpenLink      string
	InlineFastModeTitle       string
	InlineFastModeDescription string
	PreviewBrokenButton       string
}

var localized = map[string]*Strings{
	"en": {
		Welcome:                     "Welcome!\n\nThis bot blah-blah-blah blah-blah-blah\n\nBlah-blah-blah blah-blah-blah\n\nSend /help to see all commands",
		WelcomeBack:                 "Welcome back! Send /help if you don't remember how to use this bot",
		Help:                        "Blah-blah-blah blah-blah-blah blah-blah-blah blah-blah-blah.\n\nBlah-blah-blah blah-blah-blah blah-blah-blah blah-blah-blah!",
		Settings:                    "Your current settings:",
		SettingsFastMode:            "⏱️ Fast mode",
		SettingsFastModeEnabled:     "ON",
		SettingsFastModeDisabled:    "OFF",
		SettingsCaption:             "📝 Caption: ",
		SettingsCaptionVideoOnly:    "video only",
		SettingsCaptionVideoAndLink: "video and link",
		SettingsCaptionFull:         "full",
		SettingsLanguage:            "🇬🇧 Language",
		SettingsBackButton:          "⬅️ Back",
		Farewell:                    "Your data will be deleted from database, bot will forget you. Are you sure you want to proceed?",
		FarewellConfirm:             "Yes",
		FarewellCancel:              "Cancel",
		FarewellInProgress:          "Deleting data...",
		FarewellDone:                "Data deletion is complete.",
		FarewellComplete:            "Your data was deleted",
		FarewellCanceled:            "Canceled",
		FarewellGoodbye:             "Data deletion is complete.\n\nFarewell, cruel human",
		RateLimited:                 "Too many requests, please try again later",
		SettingsError:               "Failed to get settings",
		ReplyError:                  "Failed to reply",
		RequestID:                   "Request ID",
		NoLink:                      "No link found",
		UnsupportedLink:             "Unsupported link, use /help to see supported list",
		Downloading:                 "⏳ Downloading...",
		FailedToProcessLink:         "Failed to process link",
		InlineSendingHint:           "🐾 Tap to send",
		InlineResultPlaceholder:     "%s\n\n%s\n\n⏳ Message will be updated upon processing completion\n\n%s",
		InlineButtonDownloading:     "⏳ Downloading...",
		InlineButtonOpenLink:        "🌐 Open link",
		InlineFastModeTitle:         "Download video",
		InlineFastModeDescription:   "Tap to send — video will appear shortly",
		PreviewBrokenButton:         "Preview not working?",
	},
	"ru": {
		Welcome:                     "Привет!\n\nЭтот бот бла-бла-бла бла-бла-бла\n\nБла-бла-бла бла-бла-бла бла-бла-бла\n\nОтправь /help для списка команд",
		WelcomeBack:                 "С возвращением! Отправь /help если не помнишь как пользоваться ботом",
		Help:                        "Бла-бла-бла бла-бла-бла бла-бла-бла бла-бла-бла.\n\nБла-бла-бла бла-бла-бла бла-бла-бла бла-бла-бла!",
		Settings:                    "Твои настройки:",
		SettingsFastMode:            "⏱️ Быстрый режим",
		SettingsFastModeEnabled:     "вкл.",
		SettingsFastModeDisabled:    "выкл.",
		SettingsCaption:             "📝 Подпись: ",
		SettingsCaptionVideoOnly:    "только видео",
		SettingsCaptionVideoAndLink: "видео и ссылка",
		SettingsCaptionFull:         "полностью",
		SettingsLanguage:            "🇷🇺 Язык",
		SettingsBackButton:          "⬅️ Назад",
		Farewell:                    "Связанная с вами информация будет удалена из базы данных, бот забудет о вас.\n\nВы уверены что хотите продолжить?",
		FarewellConfirm:             "Да",
		FarewellCancel:              "Отмена",
		FarewellInProgress:          "Удаление данных...",
		FarewellDone:                "Удаление данных завершено.",
		FarewellComplete:            "Ваши данные были удалены",
		FarewellCanceled:            "Отменено",
		FarewellGoodbye:             "Удаление данных завершено.\n\nFarewell, cruel human",
		RateLimited:                 "Слишком много запросов, попробуй позже",
		SettingsError:               "Ошибка, попробуй ещё раз",
		ReplyError:                  "Не удалось ответить",
		RequestID:                   "ID запроса",
		NoLink:                      "Ссылка не найдена",
		UnsupportedLink:             "Неподдерживаемая ссылка, отправь /help для списка",
		Downloading:                 "⏳ Загружается...",
		FailedToProcessLink:         "Не удалось обработать ссылку",
		InlineSendingHint:           "🐾 Нажми для отправки",
		InlineResultPlaceholder:     "%s\n\n%s\n\n⏳ Сообщение обновится после обработки\n\n%s",
		InlineButtonDownloading:     "⏳ Загружается...",
		InlineButtonOpenLink:        "🌐 Открыть ссылку",
		InlineFastModeTitle:         "Скачать видео",
		InlineFastModeDescription:   "Нажми чтобы отправить — видео появится через некоторое время",
		PreviewBrokenButton:         "Превью не работает?",
	},
	"ua": {
		Welcome:                     "Привіт!\n\nЦей бот бла-бла-бла бла-бла-бла\n\nБла-бла-бла бла-бла-бла бла-бла-бла\n\nНадішли /help для списку команд",
		WelcomeBack:                 "З поверненням! Надішли /help якщо забув як користуватися",
		Help:                        "Бла-бла-бла бла-бла-бла бла-бла-бла бла-бла-бла.\n\nБла-бла-бла бла-бла-бла бла-бла-бла бла-бла-бла!",
		Settings:                    "Твої налаштування",
		SettingsFastMode:            "⏱️ Швидкий режим",
		SettingsFastModeEnabled:     "увімк.",
		SettingsFastModeDisabled:    "вимк.",
		SettingsCaption:             "📝 Підпис: ",
		SettingsCaptionVideoOnly:    "тільки відео",
		SettingsCaptionVideoAndLink: "відео та посилання",
		SettingsCaptionFull:         "повністю",
		SettingsLanguage:            "🇺🇦 Мова",
		SettingsBackButton:          "⬅️ Назад",
		Farewell:                    "Пов'язана з вами інформація буде видалена з бази даних, бот забуде про вас.\n\nВи впевнені, що хочете продовжити?",
		FarewellConfirm:             "Так",
		FarewellCancel:              "Скасувати",
		FarewellInProgress:          "Видалення даних...",
		FarewellDone:                "Видалення даних завершено.",
		FarewellComplete:            "Ваші дані були видалені",
		FarewellCanceled:            "Скасовано",
		FarewellGoodbye:             "Видалення даних завершено.\n\nFarewell, cruel human",
		RateLimited:                 "Забагато запитів, спробуй пізніше",
		SettingsError:               "Помилка, спробуй ще раз",
		ReplyError:                  "Не вдалося відповісти",
		RequestID:                   "ID запиту",
		NoLink:                      "Посилання не знайдено",
		UnsupportedLink:             "Непідтримуване посилання, надішли /help для списку",
		Downloading:                 "⏳ Завантажується...",
		FailedToProcessLink:         "Не вдалося обробити посилання",
		InlineSendingHint:           "🐾 Натисніть, щоб надіслати",
		InlineResultPlaceholder:     "%s\n\n%s\n\n⏳ Повідомлення оновиться після обробки\n\n%s",
		InlineButtonDownloading:     "⏳ Завантажується...",
		InlineButtonOpenLink:        "🌐 Відкрити посилання",
		InlineFastModeTitle:         "Завантажити відео",
		InlineFastModeDescription:   "Натисни щоб надіслати — відео з'явиться незабаром",
		PreviewBrokenButton:         "Прев'ю не працює?",
	},
}

func normalizeLanguageCode(languageCode string) string {
	if languageCode == "uk" {
		languageCode = "ua"
	} // Somewhy Ukrainian language code is "uk" in Telegram

	return languageCode
}

func GetLocalizedString(languageCode string) *Strings {
	languageCode = normalizeLanguageCode(languageCode)

	if m, ok := localized[languageCode]; ok {
		return m
	}

	return localized["en"]
}

func GetUserLanguageCode(ctx *ext.Context) string {
	if ctx.EffectiveUser != nil {
		return normalizeLanguageCode(ctx.EffectiveUser.LanguageCode)
	}

	if ctx.InlineQuery != nil {
		return normalizeLanguageCode(ctx.InlineQuery.From.LanguageCode)
	}

	if ctx.ChosenInlineResult != nil {
		return normalizeLanguageCode(ctx.ChosenInlineResult.From.LanguageCode)
	}

	if ctx.Update.ChosenInlineResult != nil {
		return normalizeLanguageCode(ctx.Update.ChosenInlineResult.From.LanguageCode)
	}

	return "en"
}

func Lang(languageCode string) *Strings {
	return GetLocalizedString(languageCode)
}

func LangCtx(ctx *ext.Context) *Strings {
	return GetLocalizedString(GetUserLanguageCode(ctx))
}
