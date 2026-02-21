package consumer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/Anton9372/go-lib/brokers/rabbit"
	"github.com/Anton9372/go-lib/logger"
)

type MsgHandler interface {
	Handle(ctx context.Context, msg amqp.Delivery) error
}

type Consumer struct {
	client        *rabbit.Client
	queueName     string
	consumerTag   string
	prefetchCount int

	l *slog.Logger

	msgHandler MsgHandler
}

func NewConsumer(client *rabbit.Client, queueName string, l *slog.Logger, msgHandler MsgHandler) *Consumer {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown-host"
	}
	consumerTag := fmt.Sprintf("%s-%s", queueName, hostname)

	l = l.With("queue-name", queueName, "consumer-tag", consumerTag)

	return &Consumer{
		client:        client,
		queueName:     queueName,
		consumerTag:   consumerTag,
		prefetchCount: 1,
		l:             l,
		msgHandler:    msgHandler,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		err := c.consumeLoop(ctx)
		if err != nil {
			c.l.Error("Consume loop stopped, will retry...", logger.ErrAttr(err))
		}

		if ctx.Err() != nil {
			c.l.Info("Consumer gracefully stopped")
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second): //nolint:mnd // the value is clear
		}
	}
}

func (c *Consumer) consumeLoop(ctx context.Context) error {
	ch, err := c.client.Channel()
	if err != nil {
		return fmt.Errorf("get channel: %w", err)
	}

	if err = ch.Qos(c.prefetchCount, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}

	msgs, err := ch.Consume(
		c.queueName,
		c.consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("register consume: %w", err)
	}

	c.l.Info("Consumer started listening")

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("amqp messages channel closed")
			}

			c.consumeMsg(ctx, msg)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Consumer) consumeMsg(ctx context.Context, msg amqp.Delivery) {
	const empty = "empty"

	msgID := msg.MessageId
	if msgID == "" {
		msgID = empty
	}
	corrID := msg.CorrelationId
	if corrID == "" {
		corrID = empty
	}

	msgLogger := c.l.With(
		slog.String("message-id", msgID),
		slog.String("correlation-id", corrID),
		slog.String("routing-key", msg.RoutingKey),
	)

	msgCtx := logger.ContextWithLogger(ctx, msgLogger)

	err := c.msgHandler.Handle(msgCtx, msg)

	if err != nil {
		c.l.Error("Unable to handle message, nacking", logger.ErrAttr(err))
		if nackErr := msg.Nack(false, true); nackErr != nil {
			c.l.Error("Unable to NACK RabbitMQ message", logger.ErrAttr(nackErr))
		}
	} else {
		if ackErr := msg.Ack(false); ackErr != nil {
			c.l.Error("Unable to ACK RabbitMQ message", logger.ErrAttr(ackErr))
		}
	}
}
