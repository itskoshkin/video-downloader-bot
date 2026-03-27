package models

import (
	"time"
	s "video-downloader-bot/internal/telegram/strings"

	"gorm.io/gorm"
)

type User struct {
	ID                  uint  `gorm:"primaryKey"`
	TelegramID          int64 `gorm:"not null"`
	Username            string
	FirstName           string
	LastName            string
	Language            string
	IsPremium           bool
	IsBot               bool
	SettingsFastMode    bool        `gorm:"default:false"`
	SettingsLanguage    string      `gorm:"default:'en'"`
	SettingsCaptionMode CaptionMode `gorm:"default:'full'"`
	ChatUseCount        int64
	InlineUseCount      int64
	TotalUseCount       int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
	LastSeenAt          time.Time
	DeletedAt           gorm.DeletedAt
}

type UsageType string

const (
	UsageChat   UsageType = "chat"
	UsageInline UsageType = "inline"
)

type CaptionMode string

const (
	CaptionVideoOnly    CaptionMode = "video"
	CaptionVideoAndLink CaptionMode = "link"
	CaptionFullDetails  CaptionMode = "full" // icon + author + description + link
)

func (m CaptionMode) Next() CaptionMode {
	switch m {
	case CaptionVideoOnly:
		return CaptionVideoAndLink
	case CaptionVideoAndLink:
		return CaptionFullDetails
	default:
		return CaptionVideoOnly
	}
}

func (m CaptionMode) Label(languageCode string) string {
	switch m {
	case CaptionVideoOnly:
		return s.Lang(languageCode).SettingsCaptionVideoOnly
	case CaptionVideoAndLink:
		return s.Lang(languageCode).SettingsCaptionVideoAndLink
	default:
		return s.Lang(languageCode).SettingsCaptionFull
	}
}

func (u *User) Lang() string {
	return u.SettingsLanguage
}
