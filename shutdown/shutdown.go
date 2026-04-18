package shutdown

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/Anton9372/go-lib/logger"
)

type CloseStage []CloseFunc

type CloseFunc struct {
	Name string
	F    func(context.Context) error
}

func CloseConnections(stages []CloseStage, l *slog.Logger, timeout time.Duration) {
	l.Warn("Closing connections", slog.Duration("timeout", timeout))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})

	go func() {
		defer close(done)

		for i := len(stages) - 1; i >= 0; i-- {
			stage := stages[i]
			if len(stage) == 0 {
				continue
			}

			l.Debug("Starting shutdown stage", slog.Int("stage_index", i), slog.Int("closers", len(stage)))

			var wg sync.WaitGroup
			wg.Add(len(stage))

			for _, closeFn := range stage {
				go func(fn CloseFunc) {
					defer wg.Done()

					if fn.F == nil {
						l.Warn("Close function is nil, skipping", slog.String("name", fn.Name))
						return
					}

					if err := fn.F(ctx); err != nil {
						l.Error("Closing connection", slog.String("name", fn.Name), logger.ErrAttr(err))
					}
				}(closeFn)
			}

			stageDone := make(chan struct{})
			go func() {
				wg.Wait()
				close(stageDone)
			}()

			select {
			case <-stageDone:
			case <-ctx.Done():
				l.Error("Stage shutdown interrupted by timeout", slog.Int("stage_index", i))
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		l.Error("Context deadline exceeded before shutdown completed")
	case <-done:
		l.Info("All connections closed successfully")
	}

	l.Info("App terminated")
}
