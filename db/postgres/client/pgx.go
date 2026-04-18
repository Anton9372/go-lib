package client

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Anton9372/go-lib/shutdown"
)

const ErrCodeUniqueViolation = "23505"

type Config struct {
	Host     string `env:"POSTGRES_HOST"`
	Port     string `env:"POSTGRES_PORT"`
	Database string `env:"POSTGRES_DATABASE"`
	Username string `env:"POSTGRES_USERNAME"`
	Password string `env:"POSTGRES_PASSWORD"`
}

func NewPgxPool(ctx context.Context, cfg Config) (*pgxpool.Pool, shutdown.CloseFunc, error) {
	q := url.Values{}
	q.Add("sslmode", "disable")

	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.Username, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, cfg.Port),
		Path:     cfg.Database,
		RawQuery: q.Encode(),
	}
	dsn := u.String()

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, shutdown.CloseFunc{}, fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, shutdown.CloseFunc{}, fmt.Errorf("connect to postgres: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, shutdown.CloseFunc{}, fmt.Errorf("ping postgres: %w", err)
	}

	closeFn := shutdown.CloseFunc{
		Name: "PgxClient",
		F: func(_ context.Context) error {
			pool.Close()
			return nil
		},
	}

	return pool, closeFn, nil
}
