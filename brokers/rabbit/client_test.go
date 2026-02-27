package rabbit_test

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"

	"github.com/Anton9372/go-lib/brokers/rabbit"
)

func setupRabbitContainer(ctx context.Context, t *testing.T) (*rabbitmq.RabbitMQContainer, rabbit.ClientConfig) {
	t.Helper()

	rabbitContainer, err := rabbitmq.Run(ctx, "rabbitmq:3.12-management-alpine")
	require.NoError(t, err, "failed to start rabbitmq container")

	amqpURL, err := rabbitContainer.AmqpURL(ctx)
	require.NoError(t, err, "failed to get amqp url")

	parsedURL, err := url.Parse(amqpURL)
	require.NoError(t, err)

	host, port, err := net.SplitHostPort(parsedURL.Host)
	require.NoError(t, err)

	password, _ := parsedURL.User.Password()

	cfg := rabbit.ClientConfig{
		Host:             host,
		Port:             port,
		VHost:            "/",
		Username:         parsedURL.User.Username(),
		Password:         password,
		ReconnectTimeout: 100 * time.Millisecond,
	}

	return rabbitContainer, cfg
}

//nolint:paralleltest // github workflow will cry as we are launching new container for each test
func TestClient_ConnectAndClose(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	container, cfg := setupRabbitContainer(ctx, t)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate container: %v", err)
		}
	}()

	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	client := rabbit.NewClient(cfg, l)

	closeFunc, err := client.Connect(ctx)
	require.NoError(t, err)
	require.NotNil(t, closeFunc)

	ch, err := client.NewChannel()
	require.NoError(t, err)
	require.NotNil(t, ch)
	assert.False(t, ch.IsClosed(), "channel should be open")

	err = closeFunc(ctx)
	require.NoError(t, err)

	_, err = client.NewChannel()
	assert.ErrorContains(t, err, "RabbitMQ connection is not ready")
}

//nolint:paralleltest // github workflow will cry as we are launching new container for each test
func TestClient_ReconnectOnNetworkFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	container, cfg := setupRabbitContainer(ctx, t)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate container: %v", err)
		}
	}()

	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	client := rabbit.NewClient(cfg, l)

	closeFunc, err := client.Connect(ctx)
	require.NoError(t, err)
	defer func() {
		if err = closeFunc(ctx); err != nil {
			t.Errorf("Close func: %v", err)
		}
	}()

	_, err = client.NewChannel()
	require.NoError(t, err)

	_, _, err = container.Exec(ctx, []string{"rabbitmqctl", "stop_app"})
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	_, err = client.NewChannel()
	require.ErrorContains(t, err, "RabbitMQ connection is not ready", "client should block channel access during outage")

	_, _, err = container.Exec(ctx, []string{"rabbitmqctl", "start_app"})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ch, err := client.NewChannel()
		if err != nil {
			return false
		}
		err = ch.ExchangeDeclare("test_recovery_exchange", "fanout", false, true, false, false, nil)
		return err == nil
	}, 10*time.Second, 100*time.Millisecond, "Client failed to auto-reconnect to RabbitMQ")
}

//nolint:paralleltest // github workflow will cry as we are launching new container for each test
func TestClient_ConcurrentAccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	container, cfg := setupRabbitContainer(ctx, t)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate container: %v", err)
		}
	}()

	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	client := rabbit.NewClient(cfg, l)

	closeFunc, err := client.Connect(ctx)
	require.NoError(t, err)
	defer func() {
		if err = closeFunc(ctx); err != nil {
			t.Errorf("Close func: %v", err)
		}
	}()

	var wg sync.WaitGroup
	workers := 50

	for range workers {
		wg.Go(func() {
			for range 100 {
				_, _ = client.NewChannel()
				time.Sleep(1 * time.Millisecond)
			}
		})
	}

	_, _, err = container.Exec(ctx, []string{"rabbitmqctl", "close_all_connections", "Test chaos"})
	require.NoError(t, err)

	wg.Wait()
}

//nolint:paralleltest // github workflow will cry as we are launching new container for each test
func TestClient_WithChannel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	container, cfg := setupRabbitContainer(ctx, t)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate container: %v", err)
		}
	}()

	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	client := rabbit.NewClient(cfg, l)

	closeFunc, err := client.Connect(ctx)
	require.NoError(t, err)
	defer func() {
		if err = closeFunc(ctx); err != nil {
			t.Errorf("Close func: %v", err)
		}
	}()

	t.Run("success and channel closed", func(t *testing.T) {
		var capturedCh *amqp.Channel

		err = client.WithChannel(func(ch *amqp.Channel) error {
			capturedCh = ch

			assert.False(t, ch.IsClosed(), "channel should be open inside callback")

			_, declareErr := ch.QueueDeclare("test_with_channel_queue", false, true, false, false, nil)
			return declareErr
		})

		require.NoError(t, err)
		require.NotNil(t, capturedCh)

		assert.True(t, capturedCh.IsClosed(), "channel MUST be closed after WithChannel returns")
	})

	t.Run("error propagation", func(t *testing.T) {
		expectedErr := errors.New("custom business error")
		var capturedCh *amqp.Channel

		err = client.WithChannel(func(ch *amqp.Channel) error {
			capturedCh = ch
			return expectedErr
		})

		require.ErrorIs(t, err, expectedErr)

		require.NotNil(t, capturedCh)
		assert.True(t, capturedCh.IsClosed(), "channel MUST be closed even if callback returned an error")
	})
}
