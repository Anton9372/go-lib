package redis

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/Anton9372/go-lib/logger"
	"github.com/Anton9372/go-lib/shutdown"
)

type Config struct {
	Host     string `env:"REDIS_HOST"     yaml:"host"`
	Port     string `env:"REDIS_PORT"     yaml:"port"`
	Password string `env:"REDIS_PASSWORD" yaml:"password"`
	Database int    `env:"REDIS_DATABASE" yaml:"database"`
}

func New(ctx context.Context, cfg Config, l *slog.Logger) (*redis.Client, shutdown.CloseFunc, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		if closeErr := client.Close(); closeErr != nil {
			l.Error("Unable to close redis client", logger.ErrAttr(err))
		}
		return nil, shutdown.CloseFunc{}, fmt.Errorf("ping redis: %w", err)
	}

	closeFn := shutdown.CloseFunc{
		Name: "RedisClient",
		F: func(_ context.Context) error {
			return client.Close()
		},
	}

	return client, closeFn, nil
}
