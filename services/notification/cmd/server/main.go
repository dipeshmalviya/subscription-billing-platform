package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	nkafka "github.com/dipeshmalviya/subscription-billing-platform/notification/internal/kafka"
	"github.com/dipeshmalviya/subscription-billing-platform/notification/internal/notifier"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	topics := []string{"subscription.created", "subscription.canceled", "payment.succeeded", "payment.failed"}

	n := notifier.NewEmailNotifier(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for _, topic := range topics {
		consumer := nkafka.NewConsumer(brokers, topic, "notification-service", n, logger)
		go func(t string) {
			logger.Info().Str("topic", t).Msg("listening for events")
			if err := consumer.Run(ctx); err != nil && ctx.Err() == nil {
				logger.Error().Err(err).Str("topic", t).Msg("consumer stopped")
			}
		}(topic)
	}

	go func() {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})
		http.ListenAndServe(":"+getEnv("PORT", "8082"), nil)
	}()

	<-ctx.Done()
	logger.Info().Msg("shutting down notification service")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}