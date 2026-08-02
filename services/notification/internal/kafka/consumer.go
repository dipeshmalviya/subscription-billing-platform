package kafka

import (
	"context"
	"encoding/json"

	"github.com/dipeshmalviya/subscription-billing-platform/notification/internal/notifier"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Consumer struct {
	reader   *kafka.Reader
	notifier *notifier.EmailNotifier
	logger   zerolog.Logger
}

func NewConsumer(brokers []string, topic, groupID string, n *notifier.EmailNotifier, logger zerolog.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	return &Consumer{reader: reader, notifier: n, logger: logger}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return err
		}

		var event Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			c.logger.Error().Err(err).Msg("failed to unmarshal event")
			continue
		}

		message := buildMessage(event.Type)
		_ = c.notifier.Send(string(msg.Key), event.Type, message)
	}
}

func buildMessage(eventType string) string {
	switch eventType {
	case "subscription.created":
		return "Your subscription is now active."
	case "subscription.canceled":
		return "Your subscription has been canceled."
	case "payment.succeeded":
		return "Your payment was processed successfully."
	case "payment.failed":
		return "Your payment could not be processed. We'll retry shortly."
	default:
		return "You have a new notification."
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}