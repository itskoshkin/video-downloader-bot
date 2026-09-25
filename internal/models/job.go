package models

import "time"

// Job is a link that is being processed right now; the row lives only until the result or an error is shown.
// A row that outlives the processing timeout means its process died (OOM, restart), so the sweeper attaches a retry button to its message.
type Job struct {
	ID              uint   `gorm:"primaryKey"`
	UserID          int64  `gorm:"not null"`
	Lang            string // Language of the retry button the sweeper attaches
	Link            string `gorm:"not null"` // Link as the user sent it, so a retry runs exactly the same request
	ChatID          int64  // Private chat: chat of the "⏳ Downloading..." status message
	MessageID       int64  // Private chat: the "⏳ Downloading..." status message
	InlineMessageID string // Inline mode: the placeholder message
	StaleAt         *time.Time
	CreatedAt       time.Time `gorm:"index"`
}
