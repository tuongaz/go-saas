package log

import (
	"context"
	"log/slog"
)

type ctxKey string

var traceIDKey = ctxKey("trace-id-key")

type Hook interface {
	Handle(ctx context.Context, record *slog.Record)
}

func NewHandlerWithHooks(h slog.Handler, hooks ...Hook) slog.Handler {
	return &handlerWithHooks{
		hooks:   hooks,
		handler: h,
	}
}

type handlerWithHooks struct {
	hooks   []Hook
	handler slog.Handler
}

func (h handlerWithHooks) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h handlerWithHooks) Handle(ctx context.Context, record slog.Record) error {
	for _, h := range h.hooks {
		h.Handle(ctx, &record)
	}

	return h.handler.Handle(ctx, record)
}

func (h handlerWithHooks) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewHandlerWithHooks(h.handler.WithAttrs(attrs), h.hooks...)
}

func (h handlerWithHooks) WithGroup(name string) slog.Handler {
	return NewHandlerWithHooks(h.handler.WithGroup(name), h.hooks...)
}

type HookFn func(ctx context.Context, record *slog.Record)

func (fn HookFn) Handle(ctx context.Context, record *slog.Record) {
	fn(ctx, record)
}

var TracingHook = HookFn(func(ctx context.Context, record *slog.Record) {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		record.AddAttrs(slog.String("trace-id", id))
	}
})
