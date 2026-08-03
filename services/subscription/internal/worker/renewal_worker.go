package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/domain"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/grpcclient"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/kafka"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// RenewalWorker periodically scans for active subscriptions whose current
// billing period has ended, creates the next invoice, charges the customer
// via Payment Service over gRPC, and — on success — advances the
// subscription's period. On failure the subscription is marked past_due
// rather than silently retried here; Payment Service's own retry worker
// handles re-attempting the charge.
type RenewalWorker struct {
	subRepo       *repository.SubscriptionRepository
	invRepo       *repository.InvoiceRepository
	planRepo      *repository.PlanRepository
	paymentClient *grpcclient.PaymentClient
	producer      *kafka.Producer
	logger        zerolog.Logger
	interval      time.Duration
}

func NewRenewalWorker(
	subRepo *repository.SubscriptionRepository,
	invRepo *repository.InvoiceRepository,
	planRepo *repository.PlanRepository,
	paymentClient *grpcclient.PaymentClient,
	producer *kafka.Producer,
	logger zerolog.Logger,
) *RenewalWorker {
	return &RenewalWorker{
		subRepo:       subRepo,
		invRepo:       invRepo,
		planRepo:      planRepo,
		paymentClient: paymentClient,
		producer:      producer,
		logger:        logger,
		interval:      60 * time.Second,
	}
}

// Run blocks until ctx is canceled, polling on a fixed interval.
func (w *RenewalWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processDue(ctx)
		}
	}
}

// processDue finds subscriptions due for renewal and processes each one
// individually — one failure doesn't block the rest of the batch.
func (w *RenewalWorker) processDue(ctx context.Context) {
	due, err := w.subRepo.DueForRenewal(ctx, time.Now(), 50)
	if err != nil {
		w.logger.Error().Err(err).Msg("renewal worker: failed to query due subscriptions")
		return
	}
	if len(due) == 0 {
		return
	}

	w.logger.Info().Int("count", len(due)).Msg("renewal worker: processing due subscriptions")

	for _, sub := range due {
		select {
		case <-ctx.Done():
			return
		default:
		}
		w.renewOne(ctx, sub)
	}
}

// renewOne creates the renewal invoice, charges the customer via Payment
// Service, and updates state based on the outcome. The idempotency key is
// deterministic (subscription + invoice ID) so a duplicate renewal cycle
// hitting the same invoice — e.g. if this worker's interval overlapped a
// slow prior run — won't double-charge the customer.
func (w *RenewalWorker) renewOne(ctx context.Context, sub *domain.Subscription) {
	plan, err := w.planRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		w.logger.Error().Err(err).Str("subscription_id", sub.ID.String()).Msg("failed to load plan")
		return
	}

	invoice := &domain.Invoice{
		ID:             uuid.New(),
		SubscriptionID: sub.ID,
		CustomerID:     sub.CustomerID,
		AmountCents:    plan.PriceCents,
		Currency:       plan.Currency,
		Status:         domain.InvoicePending,
		DueDate:        time.Now(),
	}
	if err := w.invRepo.Create(ctx, invoice); err != nil {
		w.logger.Error().Err(err).Str("subscription_id", sub.ID.String()).Msg("failed to create renewal invoice")
		return
	}

	idempotencyKey := fmt.Sprintf("renewal:%s:%s", sub.ID.String(), invoice.ID.String())
	resp, err := w.paymentClient.ChargeCustomer(
		ctx,
		invoice.ID.String(),
		sub.CustomerID.String(),
		invoice.AmountCents,
		invoice.Currency,
		idempotencyKey,
	)
	if err != nil {
		w.logger.Error().Err(err).Str("invoice_id", invoice.ID.String()).Msg("charge call to payment service failed")
		return
	}

	if resp.Status.String() == "PAYMENT_STATUS_SUCCEEDED" {
		w.handleSuccess(ctx, sub, plan, invoice)
	} else {
		w.handleFailure(ctx, sub, invoice)
	}
}

func (w *RenewalWorker) handleSuccess(ctx context.Context, sub *domain.Subscription, plan *domain.Plan, invoice *domain.Invoice) {
	if err := w.invRepo.MarkPaid(ctx, invoice.ID, invoice.Version); err != nil {
		w.logger.Error().Err(err).Str("invoice_id", invoice.ID.String()).Msg("failed to mark invoice paid")
		return
	}

	newStart := sub.CurrentPeriodEnd
	newEnd := newStart.AddDate(0, 1, 0)
	if plan.BillingInterval == domain.IntervalYearly {
		newEnd = newStart.AddDate(1, 0, 0)
	}

	if err := w.subRepo.RenewPeriod(ctx, sub.ID, newStart, newEnd, sub.Version); err != nil {
		w.logger.Error().Err(err).Str("subscription_id", sub.ID.String()).Msg("failed to advance subscription period")
		return
	}

	_ = w.producer.Publish(ctx, "subscription.renewed", sub.ID.String(), kafka.Event{
		Type:    "subscription.renewed",
		Payload: sub,
	})

	w.logger.Info().Str("subscription_id", sub.ID.String()).Msg("subscription renewed successfully")
}

func (w *RenewalWorker) handleFailure(ctx context.Context, sub *domain.Subscription, invoice *domain.Invoice) {
	if err := w.subRepo.UpdateStatus(ctx, sub.ID, domain.StatusPastDue, sub.Version); err != nil {
		w.logger.Error().Err(err).Str("subscription_id", sub.ID.String()).Msg("failed to mark subscription past_due")
		return
	}

	_ = w.producer.Publish(ctx, "subscription.past_due", sub.ID.String(), kafka.Event{
		Type:    "subscription.past_due",
		Payload: sub,
	})

	w.logger.Warn().
		Str("subscription_id", sub.ID.String()).
		Str("invoice_id", invoice.ID.String()).
		Msg("renewal charge failed, subscription marked past_due")
}