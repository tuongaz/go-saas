package log

import (
	"context"
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.New(
		NewHandlerWithHooks(
			slog.NewJSONHandler(os.Stderr, nil),
			TracingHook,
		),
	)
}

func Default() *slog.Logger {
	return defaultLogger
}

func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

func Panic(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
	panic(msg)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	defaultLogger.ErrorContext(ctx, msg, args...)
}

func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	defaultLogger.InfoContext(ctx, msg, args...)
}

func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	defaultLogger.DebugContext(ctx, msg, args...)
}

func ErrorAttr(err error) slog.Attr {
	return slog.Any("error", err)
}
