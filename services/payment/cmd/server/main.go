package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	paymentv1 "github.com/dipeshmalviya/subscription-billing-platform/gen/go/payment/v1"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/config"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/gateway"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/grpcserver"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/kafka"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/observability"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/repository"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/service"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/worker"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	logger := observability.InitLogger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	var paymentRepo *repository.PaymentRepository
	var attemptRepo *repository.PaymentAttemptRepository

	pool, err := pgxpool.New(context.Background(), cfg.PostgresURI)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to initialize postgres pool")
	} else {
		defer pool.Close()
		if err := pool.Ping(context.Background()); err != nil {
			logger.Warn().Err(err).Msg("postgres ping failed; continuing without repository initialization")
		} else {
			logger.Info().Msg("connected to postgres")
			paymentRepo = repository.NewPaymentRepository(pool)
			attemptRepo = repository.NewPaymentAttemptRepository(pool)
		}
	}

	mockGateway := gateway.NewMockGateway()

	producer := kafka.NewProducer(strings.Split(cfg.KafkaBrokers, ","))
	defer producer.Close()

	paymentService := service.NewPaymentService(paymentRepo, attemptRepo, mockGateway, producer)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	retryWorker := worker.NewRetryWorker(paymentRepo, paymentService, logger)
	go retryWorker.Run(ctx)
	logger.Info().Msg("payment retry worker started")

	grpcServer := grpc.NewServer()
	paymentv1.RegisterPaymentServiceServer(grpcServer, grpcserver.NewPaymentServer(paymentService))

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		logger.Fatal().Err(err).Str("port", cfg.Port).Msg("failed to listen")
	}

	go func() {
		logger.Info().Str("port", cfg.Port).Msg("payment service gRPC server starting")
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal().Err(err).Msg("grpc server failed")
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", observability.HealthCheckHandler)
	mux.Handle("/metrics", promhttp.Handler())

	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	go func() {
		logger.Info().Str("port", cfg.HTTPPort).Msg("payment service health/metrics server starting")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("health/metrics server failed")
		}
	}()

	<-ctx.Done()
	logger.Info().Msg("shutdown signal received, shutting down payment service")

	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(context.Background())

	os.Exit(0)
}
