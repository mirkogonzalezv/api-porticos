package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey string

const requestIDKey ctxKey = "request_id"

func WithRequestID(ctx context.Context, id string) context.Context {
	if ctx == nil {
		return context.WithValue(context.Background(), requestIDKey, id)
	}
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	raw := ctx.Value(requestIDKey)
	if v, ok := raw.(string); ok {
		return v
	}
	return ""
}

func FromContext(ctx context.Context) *zap.Logger {
	if id := RequestIDFrom(ctx); id != "" {
		return L().With(zap.String("request_id", id))
	}
	return L()
}
