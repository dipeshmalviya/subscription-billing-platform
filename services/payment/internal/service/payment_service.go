package service

import (
	"context"
	"errors"

	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/domain"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/gateway"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/kafka"
		"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/observability"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/repository"
	"github.com/google/uuid"
)

type PaymentService struct {
	paymentRepo *repository.PaymentRepository
	attemptRepo *repository.PaymentAttemptRepository
	gw          *gateway.MockGateway
	producer    *kafka.Producer
}

func NewPaymentService(
	paymentRepo *repository.PaymentRepository,
	attemptRepo *repository.PaymentAttemptRepository,
	gw *gateway.MockGateway,
	producer *kafka.Producer,
) *PaymentService {
	return &PaymentService{paymentRepo: paymentRepo, attemptRepo: attemptRepo, gw: gw, producer: producer}
}

func (s *PaymentService) Refund(ctx context.Context, paymentID uuid.UUID, reason string) (*domain.Payment, error) {
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if err := s.gw.Refund(ctx, payment.ProviderRef); err != nil {
		return nil, err
	}
	if err := s.paymentRepo.UpdateResult(ctx, payment.ID, domain.PaymentRefunded, payment.ProviderRef, payment.Version); err != nil {
		return nil, err
	}
	payment.Status = domain.PaymentRefunded

	observability.RefundsTotal.WithLabelValues("succeeded").Inc()
	_ = s.producer.Publish(ctx, "payment.refunded", payment.ID.String(), kafka.Event{
		Type: "payment.refunded", Payload: payment,
	})
	return payment, nil
}

// ChargeCustomer is idempotent: calling it twice with the same idempotencyKey
// returns the existing payment instead of charging again.
func (s *PaymentService) ChargeCustomer(ctx context.Context, invoiceID, customerID uuid.UUID, amountCents int64, currency, idempotencyKey string) (*domain.Payment, error) {
	existing, err := s.paymentRepo.GetByIdempotencyKey(ctx, idempotencyKey)
	if err == nil {
		return existing, nil // already processed — return prior result
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	payment := &domain.Payment{
		ID:             uuid.New(),
		InvoiceID:      invoiceID,
		CustomerID:     customerID,
		AmountCents:    amountCents,
		Currency:       currency,
		Status:         domain.PaymentPending,
		IdempotencyKey: idempotencyKey,
	}
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	result, chargeErr := s.gw.Charge(ctx, amountCents, currency)

	attempt := &domain.PaymentAttempt{
		ID:            uuid.New(),
		PaymentID:     payment.ID,
		AttemptNumber: payment.AttemptCount + 1,
	}

	var newStatus domain.PaymentStatus
	if chargeErr != nil {
		newStatus = domain.PaymentFailed
		attempt.Status = domain.PaymentFailed
		attempt.ErrorMessage = chargeErr.Error()
	} else {
		newStatus = domain.PaymentSucceeded
		attempt.Status = domain.PaymentSucceeded
		payment.ProviderRef = result.ProviderRef
	}

	_ = s.attemptRepo.Record(ctx, attempt)

	if err := s.paymentRepo.UpdateResult(ctx, payment.ID, newStatus, payment.ProviderRef, payment.Version); err != nil {
		return nil, err
	}
	payment.Status = newStatus

	eventType := "payment.succeeded"
	if newStatus == domain.PaymentFailed {
		eventType = "payment.failed"
	}
	_ = s.producer.Publish(ctx, eventType, payment.ID.String(), kafka.Event{
		Type:    eventType,
		Payload: payment,
	})

	return payment, nil
}

func (s *PaymentService) GetStatus(ctx context.Context, invoiceID uuid.UUID) (*domain.Payment, error) {
	return s.paymentRepo.GetByInvoiceID(ctx, invoiceID)
}