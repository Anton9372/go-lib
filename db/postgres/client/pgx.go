package client

import (
	"context"
	"fmt"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
)

const ErrCodeUniqueViolation = "23505"

type Config struct {
	Host     string `env:"POSTGRES_HOST"`
	Port     string `env:"POSTGRES_PORT"`
	Database string `env:"POSTGRES_DATABASE"`
	Username string `env:"POSTGRES_USERNAME"`
	Password string `env:"POSTGRES_PASSWORD"`
}

func NewPgxPool(ctx context.Context, cfg Config) (*pgxpool.Pool, func(context.Context) error, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.Username,
		cfg.Password,
		net.JoinHostPort(cfg.Host, cfg.Port),
		cfg.Database,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to postgres: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("ping postgres: %w", err)
	}

	shutdown := func(_ context.Context) error {
		pool.Close()

		return nil
	}

	return pool, shutdown, nil
}
