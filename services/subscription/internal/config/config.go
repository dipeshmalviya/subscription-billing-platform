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
		PostgresURI:        getEnv("POSTGRES_URI", "postgres://malvi:23Nov2004@postgres:5432/subscription_db?sslmode=disable"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		KafkaBrokers:       getEnv("KAFKA_BROKERS", "localhost:9092"),
		SecretKey:          getEnv("SECRET_KEY", "+YPqkUHs4G3DXUYKep5cVibZQvYyCtU1kBY/L29WsMQ="),
		SecretRefreshKey:   getEnv("SECRET_REFRESH_KEY", "o4LojczaKQS15LINq7oDnpN7ui"),
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
