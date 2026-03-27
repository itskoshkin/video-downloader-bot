package telegram

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/choseninlineresult"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/inlinequery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"

	"video-downloader-bot/internal/telegram/middlewares/requests"
)

func (b *Bot) requestIDMiddleware(_ *gotgbot.Bot, ctx *ext.Context) error {
	ctx.Data = map[string]any{"request_id": req.NewRequestID()}
	return nil
}

func (b *Bot) registerMiddleware() {
	b.dispatcher.AddHandlerToGroup(handlers.NewMessage(message.All, b.requestIDMiddleware), -1) // Default group in AddHandler() is 0, so -1
	b.dispatcher.AddHandlerToGroup(handlers.NewInlineQuery(inlinequery.All, b.requestIDMiddleware), -1)
	b.dispatcher.AddHandlerToGroup(handlers.NewChosenInlineResult(choseninlineresult.All, b.requestIDMiddleware), -1)
	b.dispatcher.AddHandlerToGroup(handlers.NewCallback(nil, b.requestIDMiddleware), -1)
}
