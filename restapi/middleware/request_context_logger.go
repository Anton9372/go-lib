package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Anton9372/go-lib/logger"
	"github.com/Anton9372/go-lib/trace"
)

const (
	headerCorrelationID = "X-Correlation-ID"
)

func RequestContextLogger(baseLogger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := uuid.New().String()
		l := baseLogger.With(slog.String("request-id", reqID))

		corrID := c.GetHeader(headerCorrelationID)
		if corrID != "" {
			corrID = uuid.New().String()
		}

		l = l.With(slog.String("correlation-id", corrID))
		c.Header(headerCorrelationID, corrID)

		l.Debug("Incoming HTTP request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)

		ctx := logger.ContextWithLogger(c.Request.Context(), l)
		ctx = trace.ContextWithCorrelationID(ctx, corrID)

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
