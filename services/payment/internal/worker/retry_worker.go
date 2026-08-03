package worker

import (
	"context"
	"time"

	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/domain"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/observability"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/repository"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/service"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// maxRetryAttempts caps how many times a single payment will be retried
// before the retry worker gives up on it — prevents infinite retry loops
// on a payment that's genuinely going to keep failing (e.g. expired card).
const maxRetryAttempts = 5

// RetryWorker periodically scans for failed payments under the retry
// threshold and re-attempts them through the same PaymentService used by
// the gRPC handler — so retries go through identical business logic
// (attempt logging, event publishing, metrics) as a first-time charge.
type RetryWorker struct {
	paymentRepo *repository.PaymentRepository
	svc         *service.PaymentService
	logger      zerolog.Logger
	interval    time.Duration
}

func NewRetryWorker(paymentRepo *repository.PaymentRepository, svc *service.PaymentService, logger zerolog.Logger) *RetryWorker {
	return &RetryWorker{
		paymentRepo: paymentRepo,
		svc:         svc,
		logger:      logger,
		interval:    30 * time.Second,
	}
}

// Run blocks until ctx is canceled, polling on a fixed interval.
func (w *RetryWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

// processBatch fetches failed payments eligible for retry and attempts
// each one, with a short pause between attempts to avoid hammering the
// gateway if a burst of payments failed together.
func (w *RetryWorker) processBatch(ctx context.Context) {
	failed, err := w.paymentRepo.ListFailedForRetry(ctx, maxRetryAttempts, 50)
	if err != nil {
		w.logger.Error().Err(err).Msg("retry worker: failed to list payments")
		return
	}

	observability.FailedPaymentsPendingRetry.Set(float64(len(failed)))

	if len(failed) == 0 {
		return
	}

	w.logger.Info().Int("count", len(failed)).Msg("retry worker: retrying failed payments")

	for _, p := range failed {
		select {
		case <-ctx.Done():
			return
		default:
		}

		w.retryOne(ctx, p)
		time.Sleep(200 * time.Millisecond)
	}
}

// retryOne re-charges a single failed payment. A new idempotency key is
// generated per retry attempt — this retry is genuinely a new charge
// attempt against the gateway, distinct from the original. The payment's
// existing row is updated in place via PaymentService.ChargeCustomer's
// normal flow rather than creating a duplicate payment record.
func (w *RetryWorker) retryOne(ctx context.Context, p *domain.Payment) {
	observability.RetryAttemptsTotal.Inc()

	newKey := p.IdempotencyKey + ":retry:" + uuid.New().String()[:8]

	result, err := w.svc.ChargeCustomer(ctx, p.InvoiceID, p.CustomerID, p.AmountCents, p.Currency, newKey)
	if err != nil {
		w.logger.Error().
			Err(err).
			Str("payment_id", p.ID.String()).
			Msg("retry attempt errored")
		return
	}

	w.logger.Info().
		Str("payment_id", p.ID.String()).
		Str("status", string(result.Status)).
		Int("attempt", p.AttemptCount+1).
		Msg("retry attempt completed")
}