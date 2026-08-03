package config

import (
	"os"
)

type Config struct {
	Port         string // gRPC port
	HTTPPort     string // health check + Prometheus metrics port
	PostgresURI  string
	KafkaBrokers string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:         getEnv("PORT", "9090"),
		HTTPPort:     getEnv("HTTP_PORT", "9091"),
		PostgresURI:  getEnv("POSTGRES_URI", "postgres://malvi:23Nov2004@postgres:5432/payment_db?sslmode=disable"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
