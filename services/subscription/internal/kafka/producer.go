package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer publishes subscription lifecycle events to Kafka, consumed by
// the Notification Service. Used by both the SubscriptionService (create,
// cancel) and the renewal worker (renewed).
type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
		},
	}
}

// Event is the envelope written to every topic. Type mirrors the topic name
// (e.g. "subscription.created") so consumers can branch on it without
// needing to know which topic they read it from.
type Event struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// Publish sends an event to a topic, keyed by subscriptionID so all events
// for the same subscription land on the same Kafka partition and stay
// ordered relative to each other (created before canceled, etc.).
func (p *Producer) Publish(ctx context.Context, topic string, key string, event Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}