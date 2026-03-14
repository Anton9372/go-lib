package shutdown

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/Anton9372/go-lib/logger"
)

type CloseStage []CloseFunc

type CloseFunc func(context.Context) error

func CloseConnections(stages []CloseStage, l *slog.Logger, timeout time.Duration) {
	l.Warn("Closing connections", slog.Duration("timeout", timeout))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})

	go func() {
		for i, stage := range stages {
			if len(stage) == 0 {
				continue
			}

			l.Debug("Starting shutdown stage", slog.Int("stage_index", i), slog.Int("closers", len(stage)))

			var wg sync.WaitGroup
			wg.Add(len(stage))

			for _, closeFn := range stage {
				go func(fn CloseFunc) {
					defer wg.Done()
					if err := fn(ctx); err != nil {
						l.Error("Closing connection", logger.ErrAttr(err))
					}
				}(closeFn)
			}

			wg.Wait()

			if ctx.Err() != nil {
				break
			}
		}

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
