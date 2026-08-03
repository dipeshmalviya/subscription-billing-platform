package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/auth"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/cache"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/config"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/graphql"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/grpcclient"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/kafka"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/observability"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/repository"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/service"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/worker"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger := observability.InitLogger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	var pool *pgxpool.Pool
	var customerRepo *repository.CustomerRepository
	var planRepo *repository.PlanRepository
	var subRepo *repository.SubscriptionRepository
	var invRepo *repository.InvoiceRepository

	pool, err = pgxpool.New(context.Background(), cfg.PostgresURI)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to initialize postgres pool")
	} else {
		defer pool.Close()
		if err := pool.Ping(context.Background()); err != nil {
			logger.Warn().Err(err).Msg("postgres ping failed; continuing without repository initialization")
		} else {
			logger.Info().Msg("connected to postgres")
			customerRepo = repository.NewCustomerRepository(pool)
			planRepo = repository.NewPlanRepository(pool)
			subRepo = repository.NewSubscriptionRepository(pool)
			invRepo = repository.NewInvoiceRepository(pool)
		}
	}

	redisCache := cache.NewRedisCache(cfg.RedisAddr)
	if err := redisCache.Ping(context.Background()); err != nil {
		logger.Warn().Err(err).Msg("redis ping failed — continuing without cache guarantees")
	}

	producer := kafka.NewProducer(strings.Split(cfg.KafkaBrokers, ","))
	defer producer.Close()

	paymentClient, err := grpcclient.NewPaymentClient(cfg.PaymentServiceAddr)
	if err != nil {
		logger.Warn().Err(err).Msg("payment service unavailable; continuing without gRPC client")
	} else {
		defer paymentClient.Close()
	}

	jwtManager := auth.NewJWTManager(cfg.SecretKey, cfg.SecretRefreshKey)
	authService := service.NewAuthService(customerRepo, jwtManager)
	subscriptionService := service.NewSubscriptionService(subRepo, planRepo, invRepo, redisCache, producer)
	planService := service.NewPlanService(planRepo, redisCache)

	schema := graphql.NewExecutableSchema(graphql.Config{Resolvers: &graphql.Resolver{
		CustomerRepo:        customerRepo,
		PlanRepo:            planRepo,
		SubscriptionRepo:    subRepo,
		InvoiceRepo:         invRepo,
		AuthService:         authService,
		SubscriptionService: subscriptionService,
		PlanService:         planService,
	}})
	graphqlServer := handler.NewDefaultServer(schema)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if paymentClient != nil && subRepo != nil && invRepo != nil && planRepo != nil {
		renewalWorker := worker.NewRenewalWorker(subRepo, invRepo, planRepo, paymentClient, producer, logger)
		go renewalWorker.Run(ctx)
	} else {
		logger.Warn().Msg("subscription worker disabled until repositories and payment client are available")
	}

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	r.Get("/healthz", observability.HealthCheckHandler)
	r.Handle("/metrics", promhttp.Handler())
	r.Handle("/", playground.Handler("GraphQL playground", "/query"))
	r.With(auth.Middleware(jwtManager)).Handle("/query", graphqlServer)

	logger.Info().Str("port", cfg.Port).Msg("starting subscription service")
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		logger.Fatal().Err(err).Msg("server failed")
		os.Exit(1)
	}
}
