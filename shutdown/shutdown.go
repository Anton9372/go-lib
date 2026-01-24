package shutdown

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/Anton9372/go-lib/logger"
)

type CloseFunc func(context.Context) error

func CloseConnections(closers []CloseFunc, l *slog.Logger, timeout time.Duration) {
	l.Warn("Closing connections", slog.Duration("timeout", timeout))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(closers))

	for _, closeFn := range closers {
		go func(fn func(context.Context) error) {
			defer wg.Done()

			err := fn(ctx)
			if err != nil {
				l.Error("Closing connection", logger.ErrAttr(err))
			}
		}(closeFn)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		l.Error("Context deadline exceeded before shutdown completed")
	case <-done:
		l.Info("All connections closed successfully")
	}

	l.Info("App terminated")
}
