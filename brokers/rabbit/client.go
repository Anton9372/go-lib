package rabbit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/Anton9372/go-lib/logger"
	"github.com/Anton9372/go-lib/shutdown"
)

type ClientConfig struct {
	Host     string `env:"RABBIT_HOST"`
	Port     string `env:"RABBIT_PORT"`
	VHost    string `env:"RABBIT_VHOST"`
	Username string `env:"RABBIT_USERNAME"`
	Password string `env:"RABBIT_PASSWORD"`

	ReconnectTimeout time.Duration `default:"5s" env:"RABBIT_RECONNECT_TIMEOUT"`
}

func (cfg ClientConfig) dsn() string {
	return fmt.Sprintf("amqp://%s:%s@%s/%s", cfg.Username, cfg.Password, net.JoinHostPort(cfg.Host, cfg.Port), url.PathEscape(cfg.VHost))
}

type Client struct {
	mu        sync.RWMutex
	conn      *amqp.Connection
	connected bool

	l *slog.Logger

	dsn              string
	reconnectTimeout time.Duration
}

func NewClient(cfg ClientConfig, l *slog.Logger) *Client {
	return &Client{
		l:                l,
		dsn:              cfg.dsn(),
		reconnectTimeout: cfg.ReconnectTimeout,
	}
}

func (c *Client) Connect(ctx context.Context) (shutdown.CloseFunc, error) {
	for {
		err := c.connectUnsafe()
		if err != nil {
			c.l.Error(
				"RabbitMQ initial connection failed, retrying after timeout...",
				logger.ErrAttr(err),
				slog.Duration("retry-after", c.reconnectTimeout),
			)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.reconnectTimeout):
				continue
			}
		}

		c.l.Info("Successfully connected to RabbitMQ")
		break
	}

	go c.keepConnection(ctx)

	return c.close, nil
}

func (c *Client) NewChannel() (*amqp.Channel, error) {
	c.mu.RLock()

	if !c.connected || c.conn == nil {
		c.mu.RUnlock()
		return nil, errors.New("RabbitMQ connection is not ready")
	}

	conn := c.conn
	c.mu.RUnlock()

	return conn.Channel()
}

func (c *Client) WithChannel(fn func(ch *amqp.Channel) error) error {
	ch, err := c.NewChannel()
	if err != nil {
		return fmt.Errorf("unable to get channel: %w", err)
	}

	defer func() {
		err = ch.Close()
		if err != nil && !errors.Is(err, amqp.ErrClosed) {
			c.l.Warn("Unable to close RabbitMQ channel", logger.ErrAttr(err))
		}
	}()

	return fn(ch)
}

func (c *Client) keepConnection(ctx context.Context) {
	for {
		c.monitorConnection(ctx)

		if ctx.Err() != nil {
			c.l.Warn("Context cancelled, will not reconnect to RabbitMQ")
			return
		}

		c.l.Warn("RabbitMQ connection lost. Attempting to reconnect...")

		for {
			err := c.connectUnsafe()
			if err != nil {
				c.l.Error(
					"RabbitMQ reconnection failed, retrying after timeout...",
					logger.ErrAttr(err),
					slog.Duration("retry-after", c.reconnectTimeout),
				)

				select {
				case <-ctx.Done():
					c.l.Warn("Context cancelled, will not reconnect to RabbitMQ")
					return
				case <-time.After(c.reconnectTimeout):
					continue
				}
			}

			c.l.Info("Successfully reconnected to RabbitMQ")
			break
		}
	}
}

func (c *Client) connectUnsafe() error {
	conn, err := amqp.Dial(c.dsn)
	if err != nil {
		return fmt.Errorf("amqp dial: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		if err = c.conn.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
			c.l.Warn("Unable to close old RabbitMQ connection during reconnect", logger.ErrAttr(err))
		}
	}

	c.conn = conn
	c.connected = true

	return nil
}

func (c *Client) monitorConnection(ctx context.Context) {
	c.mu.RLock()

	if c.conn == nil {
		c.mu.RUnlock()
		c.l.Warn("monitorConnection called with nil connection, forcing reconnect")
		return
	}

	connCloseCh := c.conn.NotifyClose(make(chan *amqp.Error, 1))
	c.mu.RUnlock()

	select {
	case <-ctx.Done():
		return
	case err := <-connCloseCh:
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()

		if err != nil {
			c.l.Warn("RabbitMQ connection closed unexpectedly", logger.ErrAttr(err))
			return
		}

		c.l.Info("RabbitMQ connection closed normally")
		return
	}
}

func (c *Client) close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = false

	var err error

	if c.conn != nil {
		deadline, ok := ctx.Deadline()
		if ok {
			err = c.conn.CloseDeadline(deadline)
		} else {
			err = c.conn.Close()
		}

		if err != nil && !errors.Is(err, amqp.ErrClosed) {
			return fmt.Errorf("close amqp connection: %w", err)
		}
	}

	return nil
}
