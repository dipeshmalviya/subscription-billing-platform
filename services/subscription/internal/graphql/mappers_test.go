package graphql

import (
	"testing"
	"time"

	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/domain"
	"github.com/google/uuid"
)

func TestToGraphQLPlan(t *testing.T) {
	plan := &domain.Plan{
		ID:              uuid.New(),
		Name:            "Basic",
		PriceCents:      5000,
		Currency:        "USD",
		BillingInterval: domain.IntervalMonthly,
		IsActive:        true,
	}

	gqlPlan := toGraphQLPlan(plan)
	if gqlPlan.PriceCents != 5000 {
		t.Fatalf("expected price cents 5000, got %d", gqlPlan.PriceCents)
	}
}

func TestToGraphQLInvoice(t *testing.T) {
	inv := &domain.Invoice{
		ID:          uuid.New(),
		AmountCents: 2000,
		Currency:    "USD",
		Status:      domain.InvoicePending,
		DueDate:     time.Now().UTC(),
	}

	gqlInv := toGraphQLInvoice(inv)
	if gqlInv.AmountCents != 2000 {
		t.Fatalf("expected amount cents 2000, got %d", gqlInv.AmountCents)
	}
}
