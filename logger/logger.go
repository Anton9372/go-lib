package logger

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
)

type key struct{}

func ContextWithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, key{}, l)
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(key{}).(*slog.Logger); ok {
		return l
	}

	callerInfo := "unknown"
	pc, file, line, ok := runtime.Caller(1) // get the caller one level up the stack (the function that called FromContext)
	if ok {
		fn := runtime.FuncForPC(pc)
		callerInfo = fmt.Sprintf("%s:%d (%s)", file, line, fn.Name())
	}

	l := slog.Default()
	l.Error("No logger found in context, using default", slog.String("caller", callerInfo))

	return l
}

func ErrAttr(err error) slog.Attr {
	return slog.Any("err", err)
}

func Level(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
