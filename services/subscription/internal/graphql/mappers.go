package graphql

import (
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/domain"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/graphql/model"
)

func toGraphQLCustomer(c *domain.Customer) *model.Customer {
	return &model.Customer{
		ID:        c.ID,
		Email:     c.Email,
		FullName:  c.FullName,
		Role:      model.Role(c.Role),
		CreatedAt: c.CreatedAt,
	}
}

func toGraphQLCustomers(cs []*domain.Customer) []*model.Customer {
	out := make([]*model.Customer, len(cs))
	for i, c := range cs {
		out[i] = toGraphQLCustomer(c)
	}
	return out
}

func toGraphQLPlan(p *domain.Plan) *model.Plan {
	return &model.Plan{
		ID:              p.ID,
		Name:            p.Name,
		PriceCents:      int32(p.PriceCents),
		Currency:        p.Currency,
		BillingInterval: model.BillingInterval(p.BillingInterval),
		IsActive:        p.IsActive,
	}
}

func toGraphQLPlans(ps []*domain.Plan) []*model.Plan {
	out := make([]*model.Plan, len(ps))
	for i, p := range ps {
		out[i] = toGraphQLPlan(p)
	}
	return out
}

func toGraphQLUserSubscription(s *domain.Subscription, customer *domain.Customer, plan *domain.Plan) *model.UserSubscription {
	return &model.UserSubscription{
		ID:                 s.ID,
		Customer:           toGraphQLCustomer(customer),
		Plan:               toGraphQLPlan(plan),
		Status:             model.SubscriptionStatus(s.Status),
		CurrentPeriodStart: s.CurrentPeriodStart,
		CurrentPeriodEnd:   s.CurrentPeriodEnd,
		CanceledAt:         s.CanceledAt,
	}
}

func toGraphQLUserSubscriptions(ss []*domain.Subscription, customers map[string]*domain.Customer, plans map[string]*domain.Plan) []*model.UserSubscription {
	out := make([]*model.UserSubscription, len(ss))
	for i, s := range ss {
		out[i] = toGraphQLUserSubscription(s, customers[s.CustomerID.String()], plans[s.PlanID.String()])
	}
	return out
}

func toGraphQLInvoice(inv *domain.Invoice) *model.Invoice {
	return &model.Invoice{
		ID:          inv.ID,
		AmountCents: int32(inv.AmountCents),
		Currency:    inv.Currency,
		Status:      model.InvoiceStatus(inv.Status),
		DueDate:     inv.DueDate,
		PaidAt:      inv.PaidAt,
	}
}

func toGraphQLInvoices(invs []*domain.Invoice) []*model.Invoice {
	out := make([]*model.Invoice, len(invs))
	for i, inv := range invs {
		out[i] = toGraphQLInvoice(inv)
	}
	return out
}
