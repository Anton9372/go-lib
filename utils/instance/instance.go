package instance

import (
	"log/slog"
	"os"
	"sync"

	"github.com/Anton9372/go-lib/logger"
)

const unknownHostname = "unknown-hostname"

var (
	name string
	once sync.Once
)

func GetName() string {
	once.Do(func() {
		hostname, err := os.Hostname()
		if err != nil {
			slog.Default().Error("Unable to get hostname, will use default value", logger.ErrAttr(err))
		} else if hostname == "" {
			slog.Default().Warn("Empty hostname, will use default value")
			hostname = unknownHostname
		}

		name = hostname
	})

	return name
}
