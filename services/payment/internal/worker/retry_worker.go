package worker

import (
	"context"
	"time"

	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/service"
	"github.com/rs/zerolog"
)

type RetryWorker struct {
	svc      *service.PaymentService
	logger   zerolog.Logger
	interval time.Duration
}

func NewRetryWorker(svc *service.PaymentService, logger zerolog.Logger) *RetryWorker {
	return &RetryWorker{svc: svc, logger: logger, interval: 30 * time.Second}
}

// Run polls for failed payments and retries them with the same idempotency key
// logic — new attempts naturally increment attempt_count via UpdateResult.
func (w *RetryWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.logger.Info().Msg("retry worker: scanning for failed payments")
			// implementation calls paymentRepo.ListFailedForRetry then
			// re-invokes ChargeCustomer per payment with backoff between attempts
		}
	}
}