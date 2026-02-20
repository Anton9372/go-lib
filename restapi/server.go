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
	RegisterRoutes(r *gin.Engine)
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

func NewServer(cfg ServerConfig, l *slog.Logger, handlers []Handler, middlewares []gin.HandlerFunc) (*Server, shutdown.CloseFunc) {
	l.Info("Initializing HTTP server", slog.String("host", cfg.Host), slog.String("port", cfg.Port))

	router := gin.New()

	router.Use(middlewares...)

	for _, h := range handlers {
		h.RegisterRoutes(router)
	}

	server := &http.Server{
		Addr:         net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:      router,
		WriteTimeout: cfg.WriteTimeout,
		ReadTimeout:  cfg.ReadTimeout,
	}

	closeFn := func(ctx context.Context) error {
		return server.Shutdown(ctx)
	}

	return &Server{
		server: server,
		l:      l,
	}, closeFn
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
