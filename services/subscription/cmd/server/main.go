package main

import (
	"context"
	"net/http"
	"os"

	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/config"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/observability"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := observability.InitLogger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	pool, err := pgxpool.New(context.Background(), cfg.PostgresURI)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		logger.Fatal().Err(err).Msg("postgres ping failed")
	}
	logger.Info().Msg("connected to postgres")

	r := chi.NewRouter()
	r.Get("/healthz", observability.HealthCheckHandler)

	logger.Info().Str("port", cfg.Port).Msg("starting subscription service")
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		logger.Fatal().Err(err).Msg("server failed")
		os.Exit(1)
	}
}