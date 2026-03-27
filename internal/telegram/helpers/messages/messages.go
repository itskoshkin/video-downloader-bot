package messages

import (
	"video-downloader-bot/internal/logger"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func Reply(bot *gotgbot.Bot, ctx *ext.Context, message string) *gotgbot.Message {
	msg, err := ctx.EffectiveMessage.Reply(bot, message, nil)
	if err != nil {
		logger.Error("failed to reply to message: %v", err)
		return nil
	}
	return msg
}
