package trace

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/Anton9372/go-lib/logger"
)

type correlationIDKey struct{}

func ContextWithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey{}, correlationID)
}

func CorrelationIDFromContext(ctx context.Context) string {
	val, ok := ctx.Value(correlationIDKey{}).(string)
	if ok {
		return val
	}

	callerInfo := "unknown"
	pc, file, line, ok := runtime.Caller(1) // get the caller one level up the stack (the function that called FromContext)
	if ok {
		fn := runtime.FuncForPC(pc)
		callerInfo = fmt.Sprintf("%s:%d (%s)", file, line, fn.Name())
	}

	l := logger.FromContext(ctx)
	l.Error("No correlation id found in context", slog.String("caller", callerInfo))

	return ""
}
