package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port               string
	DatabaseName       string
	PostgresURI        string
	RedisAddr          string
	KafkaBrokers       string
	SecretKey          string
	SecretRefreshKey   string
	AllowedOrigins     string
	PaymentServiceAddr string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseName:       getEnv("DATABASE_NAME", "subscription_db"),
		PostgresURI:        os.Getenv("POSTGRES_URI"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		KafkaBrokers:       getEnv("KAFKA_BROKERS", "localhost:9092"),
		SecretKey:          os.Getenv("SECRET_KEY"),
		SecretRefreshKey:   os.Getenv("SECRET_REFRESH_KEY"),
		AllowedOrigins:     getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),
		PaymentServiceAddr: getEnv("PAYMENT_SERVICE_ADDR", "localhost:9090"),
	}

	if cfg.PostgresURI == "" {
		return nil, fmt.Errorf("POSTGRES_URI is required")
	}
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("SECRET_KEY is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}