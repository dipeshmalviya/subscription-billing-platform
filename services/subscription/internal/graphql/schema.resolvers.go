package graphql

import (
	"context"

	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/auth"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/domain"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/graphql/model"
	"github.com/google/uuid"
)

// --- Query resolvers ---

func (r *queryResolver) Me(ctx context.Context) (*model.Customer, error) {
	idStr, ok := auth.CustomerIDFromContext(ctx)
	if !ok {
		return nil, auth.ErrUnauthorized
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	c, err := r.CustomerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toGraphQLCustomer(c), nil
}

func (r *queryResolver) Plans(ctx context.Context) ([]*model.Plan, error) {
	plans, err := r.SubscriptionService.GetActivePlans(ctx)
	if err != nil {
		return nil, err
	}
	return toGraphQLPlans(plans), nil
}

func (r *queryResolver) MySubscriptions(ctx context.Context) ([]*model.UserSubscription, error) {
	idStr, ok := auth.CustomerIDFromContext(ctx)
	if !ok {
		return nil, auth.ErrUnauthorized
	}
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	subs, err := r.SubscriptionRepo.ListByCustomer(ctx, customerID)
	if err != nil {
		return nil, err
	}

	customers := make(map[string]*domain.Customer, len(subs))
	plans := make(map[string]*domain.Plan, len(subs))
	for _, sub := range subs {
		if _, ok := customers[sub.CustomerID.String()]; !ok {
			customer, err := r.CustomerRepo.GetByID(ctx, sub.CustomerID)
			if err != nil {
				return nil, err
			}
			customers[sub.CustomerID.String()] = customer
		}
		if _, ok := plans[sub.PlanID.String()]; !ok {
			plan, err := r.PlanRepo.GetByID(ctx, sub.PlanID)
			if err != nil {
				return nil, err
			}
			plans[sub.PlanID.String()] = plan
		}
	}

	return toGraphQLUserSubscriptions(subs, customers, plans), nil
}

func (r *queryResolver) MyInvoices(ctx context.Context) ([]*model.Invoice, error) {
	idStr, ok := auth.CustomerIDFromContext(ctx)
	if !ok {
		return nil, auth.ErrUnauthorized
	}
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	invoices, err := r.InvoiceRepo.ListByCustomer(ctx, customerID)
	if err != nil {
		return nil, err
	}
	return toGraphQLInvoices(invoices), nil
}

func (r *queryResolver) Subscription(ctx context.Context, id uuid.UUID) (*model.UserSubscription, error) {
	sub, err := r.SubscriptionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	customer, err := r.CustomerRepo.GetByID(ctx, sub.CustomerID)
	if err != nil {
		return nil, err
	}
	plan, err := r.PlanRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		return nil, err
	}
	return toGraphQLUserSubscription(sub, customer, plan), nil
}

func (r *queryResolver) AllCustomers(ctx context.Context, limit *int32, offset *int32) ([]*model.Customer, error) {
	if err := auth.RequireRole(ctx, "admin"); err != nil {
		return nil, err
	}
	l, o := 20, 0
	if limit != nil {
		l = int(*limit)
	}
	if offset != nil {
		o = int(*offset)
	}
	customers, err := r.CustomerRepo.List(ctx, l, o)
	if err != nil {
		return nil, err
	}
	return toGraphQLCustomers(customers), nil
}

func (r *queryResolver) CustomerInvoices(ctx context.Context, customerID uuid.UUID) ([]*model.Invoice, error) {
	if err := auth.RequireRole(ctx, "admin"); err != nil {
		return nil, err
	}
	invoices, err := r.InvoiceRepo.ListByCustomer(ctx, customerID)
	if err != nil {
		return nil, err
	}
	return toGraphQLInvoices(invoices), nil
}

// --- Mutation resolvers ---

func (r *mutationResolver) Signup(ctx context.Context, input model.SignupInput) (*model.AuthPayload, error) {
	customer, access, refresh, err := r.AuthService.Signup(ctx, input.Email, input.FullName, input.Password)
	if err != nil {
		return nil, err
	}
	return &model.AuthPayload{
		AccessToken:  access,
		RefreshToken: refresh,
		Customer:     toGraphQLCustomer(customer),
	}, nil
}

func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error) {
	customer, access, refresh, err := r.AuthService.Login(ctx, input.Email, input.Password)
	if err != nil {
		return nil, err
	}
	return &model.AuthPayload{
		AccessToken:  access,
		RefreshToken: refresh,
		Customer:     toGraphQLCustomer(customer),
	}, nil
}

func (r *mutationResolver) CreateSubscription(ctx context.Context, input model.CreateSubscriptionInput) (*model.UserSubscription, error) {
	idStr, ok := auth.CustomerIDFromContext(ctx)
	if !ok {
		return nil, auth.ErrUnauthorized
	}
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	sub, err := r.SubscriptionService.CreateSubscription(ctx, customerID, input.PlanID)
	if err != nil {
		return nil, err
	}
	customer, err := r.CustomerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	plan, err := r.PlanRepo.GetByID(ctx, input.PlanID)
	if err != nil {
		return nil, err
	}
	return toGraphQLUserSubscription(sub, customer, plan), nil
}

func (r *mutationResolver) CancelSubscription(ctx context.Context, id uuid.UUID) (*model.UserSubscription, error) {
	if _, ok := auth.CustomerIDFromContext(ctx); !ok {
		return nil, auth.ErrUnauthorized
	}
	sub, err := r.SubscriptionService.CancelSubscription(ctx, id)
	if err != nil {
		return nil, err
	}
	customer, err := r.CustomerRepo.GetByID(ctx, sub.CustomerID)
	if err != nil {
		return nil, err
	}
	plan, err := r.PlanRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		return nil, err
	}
	return toGraphQLUserSubscription(sub, customer, plan), nil
}

func (r *mutationResolver) CreatePlan(ctx context.Context, input model.CreatePlanInput) (*model.Plan, error) {
	if err := auth.RequireRole(ctx, "admin"); err != nil {
		return nil, err
	}
	plan, err := r.PlanService.Create(ctx, input.Name, int64(input.PriceCents), input.Currency, domain.BillingInterval(input.BillingInterval))
	if err != nil {
		return nil, err
	}
	return toGraphQLPlan(plan), nil
}

func (r *mutationResolver) DeactivatePlan(ctx context.Context, id uuid.UUID) (*model.Plan, error) {
	if err := auth.RequireRole(ctx, "admin"); err != nil {
		return nil, err
	}
	// Deactivation logic: fetch, flip is_active, persist — add
	// PlanRepository.Deactivate(ctx, id) if not already present.
	plan, err := r.PlanRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toGraphQLPlan(plan), nil
}

func (r *Resolver) Query() QueryResolver       { return &queryResolver{r} }
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

type queryResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
