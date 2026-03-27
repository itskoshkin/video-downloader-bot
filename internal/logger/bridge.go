package logger

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
)

type bridgeHandler struct {
	attrs  []slog.Attr
	groups []string
	debug  bool
}

func NewBridgeLogger() *slog.Logger {
	return slog.New(&bridgeHandler{debug: viper.GetBool(config.TelegramBotLibDebug)})
}

func (h *bridgeHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *bridgeHandler) Handle(_ context.Context, record slog.Record) error {
	if record.Level == slog.LevelDebug && !h.debug {
		return nil
	}

	msg := record.Message
	attrs := h.formatAttrs(record)
	if attrs != "" {
		msg = msg + " " + attrs
	}

	switch {
	case record.Level >= slog.LevelError:
		Error("%s", msg)
	case record.Level >= slog.LevelWarn:
		Warn("%s", msg)
	case record.Level >= slog.LevelInfo:
		Info("%s", msg)
	default:
		Debug("%s", msg)
	}

	return nil
}

func (h *bridgeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := &bridgeHandler{
		attrs:  append([]slog.Attr{}, h.attrs...),
		groups: append([]string{}, h.groups...),
		debug:  h.debug,
	}
	next.attrs = append(next.attrs, attrs...)

	return next
}

func (h *bridgeHandler) WithGroup(name string) slog.Handler {
	next := &bridgeHandler{
		attrs:  append([]slog.Attr{}, h.attrs...),
		groups: append([]string{}, h.groups...),
		debug:  h.debug,
	}
	next.groups = append(next.groups, name)

	return next
}

func (h *bridgeHandler) formatAttrs(record slog.Record) string {
	parts := make([]string, 0, len(h.attrs)+record.NumAttrs())

	for _, attr := range h.attrs {
		if rendered := h.renderAttr(attr); rendered != "" {
			parts = append(parts, rendered)
		}
	}

	record.Attrs(func(attr slog.Attr) bool {
		if rendered := h.renderAttr(attr); rendered != "" {
			parts = append(parts, rendered)
		}
		return true
	})

	return strings.Join(parts, " ")
}

func (h *bridgeHandler) renderAttr(attr slog.Attr) string {
	attr.Value = attr.Value.Resolve()

	key := attr.Key
	if len(h.groups) > 0 {
		key = strings.Join(append(append([]string{}, h.groups...), key), ".")
	}

	if key == "" {
		return fmt.Sprint(attr.Value.Any())
	}

	return fmt.Sprintf("%s=%v", key, attr.Value.Any())
}
