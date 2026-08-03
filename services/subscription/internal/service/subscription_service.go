package service

import (
	"context"
	"time"

	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/cache"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/domain"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/kafka"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/repository"
	"github.com/google/uuid"
)

const (
	plansCacheKey        = "plans:active"
	plansCacheTTL         = 10 * time.Minute
	customerSummaryPrefix = "customer_summary:"
	cancelMaxRetries      = 3
)

// SubscriptionService holds the core business logic for creating and
// canceling subscriptions — coordinating Postgres writes, Redis cache
// invalidation, and Kafka event publishing in one place so resolvers and
// workers don't duplicate this logic.
type SubscriptionService struct {
	subRepo  *repository.SubscriptionRepository
	planRepo *repository.PlanRepository
	invRepo  *repository.InvoiceRepository
	cache    *cache.RedisCache
	producer *kafka.Producer
}

func NewSubscriptionService(
	subRepo *repository.SubscriptionRepository,
	planRepo *repository.PlanRepository,
	invRepo *repository.InvoiceRepository,
	cache *cache.RedisCache,
	producer *kafka.Producer,
) *SubscriptionService {
	return &SubscriptionService{
		subRepo:  subRepo,
		planRepo: planRepo,
		invRepo:  invRepo,
		cache:    cache,
		producer: producer,
	}
}

// GetActivePlans reads through Redis first — plans change rarely, so this
// is cached aggressively to keep the plans query cheap under load.
func (s *SubscriptionService) GetActivePlans(ctx context.Context) ([]*domain.Plan, error) {
	var plans []*domain.Plan

	hit, err := s.cache.Get(ctx, plansCacheKey, &plans)
	if err == nil && hit {
		return plans, nil
	}

	plans, err = s.planRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, plansCacheKey, plans, plansCacheTTL)
	return plans, nil
}

// CreateSubscription creates a subscription and its first invoice, then
// publishes a "subscription.created" event and invalidates the customer's
// cached summary. Subscription + invoice creation isn't wrapped in a single
// DB transaction here for simplicity — see note below on when to add one.
func (s *SubscriptionService) CreateSubscription(ctx context.Context, customerID, planID uuid.UUID) (*domain.Subscription, error) {
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)
	if plan.BillingInterval == domain.IntervalYearly {
		periodEnd = now.AddDate(1, 0, 0)
	}

	sub := &domain.Subscription{
		ID:                 uuid.New(),
		CustomerID:         customerID,
		PlanID:             planID,
		Status:             domain.StatusActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
	}

	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	invoice := &domain.Invoice{
		ID:             uuid.New(),
		SubscriptionID: sub.ID,
		CustomerID:     customerID,
		AmountCents:    plan.PriceCents,
		Currency:       plan.Currency,
		Status:         domain.InvoicePending,
		DueDate:        now,
	}
	if err := s.invRepo.Create(ctx, invoice); err != nil {
		return nil, err
	}

	_ = s.producer.Publish(ctx, "subscription.created", sub.ID.String(), kafka.Event{
		Type:      "subscription.created",
		Payload:   sub,
		Timestamp: now,
	})

	_ = s.cache.Delete(ctx, customerSummaryPrefix+customerID.String())

	return sub, nil
}

// CancelSubscription retries on optimistic-lock conflicts up to
// cancelMaxRetries times. This handles the real race between a customer
// canceling and the renewal worker touching the same row at nearly the
// same moment — without this loop, one of the two writes would silently
// fail instead of being retried against the fresh version.
func (s *SubscriptionService) CancelSubscription(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	var (
		sub *domain.Subscription
		err error
	)

	for attempt := 0; attempt < cancelMaxRetries; attempt++ {
		sub, err = s.subRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}

		err = s.subRepo.UpdateStatus(ctx, id, domain.StatusCanceled, sub.Version)
		if err == nil {
			break
		}
		if err != repository.ErrVersionConflict {
			return nil, err
		}
		// version conflict — loop again and retry against the fresh row
	}
	if err != nil {
		return nil, err
	}

	sub, err = s.subRepo.GetByID(ctx, id) // reload to get final version + canceled_at
	if err != nil {
		return nil, err
	}

	_ = s.producer.Publish(ctx, "subscription.canceled", sub.ID.String(), kafka.Event{
		Type:      "subscription.canceled",
		Payload:   sub,
		Timestamp: time.Now(),
	})

	_ = s.cache.Delete(ctx, customerSummaryPrefix+sub.CustomerID.String())

	return sub, nil
}

// GetSubscription is a simple pass-through used by the GraphQL
// `subscription(id:)` query resolver.
func (s *SubscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	return s.subRepo.GetByID(ctx, id)
}