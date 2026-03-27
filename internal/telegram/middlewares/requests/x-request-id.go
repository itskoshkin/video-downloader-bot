package req

import (
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/google/uuid"
)

type ctxKey struct{}

func NewRequestID() string {
	return uuid.NewString()
}

func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(ctxKey{}).(string); ok {
		return id
	}
	return ""
}

func FromExtContext(ctx *ext.Context) context.Context {
	id, _ := ctx.Data["request_id"].(string)
	return WithRequestID(context.Background(), id)
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}
