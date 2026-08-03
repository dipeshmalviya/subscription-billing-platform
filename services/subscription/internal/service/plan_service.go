package service

import (
	"context"

	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/cache"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/domain"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/repository"
	"github.com/google/uuid"
)

type PlanService struct {
	planRepo *repository.PlanRepository
	cache    *cache.RedisCache
}

func NewPlanService(planRepo *repository.PlanRepository, cache *cache.RedisCache) *PlanService {
	return &PlanService{planRepo: planRepo, cache: cache}
}

func (s *PlanService) Create(ctx context.Context, name string, priceCents int64, currency string, interval domain.BillingInterval) (*domain.Plan, error) {
	plan := &domain.Plan{
		ID:              uuid.New(),
		Name:            name,
		PriceCents:      priceCents,
		Currency:        currency,
		BillingInterval: interval,
		IsActive:        true,
	}
	if err := s.planRepo.Create(ctx, plan); err != nil {
		return nil, err
	}
	_ = s.cache.Delete(ctx, "plans:active") // invalidate cache so the new plan shows up
	return plan, nil
}