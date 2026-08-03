package config

import "os"

type Config struct {
	Port         string
	KafkaBrokers string
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8082"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}