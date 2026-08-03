package graphql

import (
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/repository"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/service"
)

type Resolver struct {
	CustomerRepo        *repository.CustomerRepository
	PlanRepo             *repository.PlanRepository
	SubscriptionRepo     *repository.SubscriptionRepository
	InvoiceRepo          *repository.InvoiceRepository
	AuthService          *service.AuthService
	SubscriptionService  *service.SubscriptionService
	PlanService          *service.PlanService
}