package config

import (
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
		PostgresURI:        getEnv("POSTGRES_URI", "postgres://billing:billing@postgres:5432/subscription_db?sslmode=disable"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		KafkaBrokers:       getEnv("KAFKA_BROKERS", "localhost:9092"),
		SecretKey:          getEnv("SECRET_KEY", "change_this_secret_key"),
		SecretRefreshKey:   getEnv("SECRET_REFRESH_KEY", "change_this_refresh_secret_key"),
		AllowedOrigins:     getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),
		PaymentServiceAddr: getEnv("PAYMENT_SERVICE_ADDR", "localhost:9090"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
