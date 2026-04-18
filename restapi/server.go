package restapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Anton9372/go-lib/shutdown"
)

type Handler interface {
	RegisterRoutes(r *gin.RouterGroup)
}

type ServerConfig struct {
	Host         string        `env:"HTTP_HOST"`
	Port         string        `env:"HTTP_PORT"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT"`
}

type Server struct {
	server *http.Server
	l      *slog.Logger
}

type ServerOption func(r *gin.Engine) error

func NewServer(
	cfg ServerConfig,
	l *slog.Logger,
	handlers []Handler,
	middlewares []gin.HandlerFunc,
	opts ...ServerOption,
) (*Server, shutdown.CloseFunc, error) {
	l.Info("Initializing HTTP server", slog.String("host", cfg.Host), slog.String("port", cfg.Port))

	router := gin.New()

	for _, opt := range opts {
		if err := opt(router); err != nil {
			return nil, shutdown.CloseFunc{}, fmt.Errorf("apply server option: %w", err)
		}
	}

	router.Use(middlewares...)

	api := router.Group("/api")

	for _, h := range handlers {
		h.RegisterRoutes(api)
	}

	server := &http.Server{
		Addr:         net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:      router,
		WriteTimeout: cfg.WriteTimeout,
		ReadTimeout:  cfg.ReadTimeout,
	}

	closeFn := shutdown.CloseFunc{
		Name: "HTTPServer",
		F: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	}

	return &Server{
		server: server,
		l:      l,
	}, closeFn, nil
}

func WithTrustedProxies(proxies []string) ServerOption {
	return func(r *gin.Engine) error {
		if err := r.SetTrustedProxies(proxies); err != nil {
			return fmt.Errorf("set trusted proxies: %w", err)
		}
		return nil
	}
}

func (s *Server) Run() error {
	if err := s.server.ListenAndServe(); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			s.l.Warn("HTTP server shutdowns")
		default:
			return fmt.Errorf("lister and serve: %w", err)
		}
	}

	return nil
}
