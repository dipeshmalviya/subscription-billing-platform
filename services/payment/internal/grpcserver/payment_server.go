package grpcserver

import (
	"context"
	"time"

	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/domain"
	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/service"
	paymentv1 "github.com/dipeshmalviya/subscription-billing-platform/gen/go/payment/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	svc *service.PaymentService
}

func NewPaymentServer(svc *service.PaymentService) *PaymentServer {
	return &PaymentServer{svc: svc}
}

func (s *PaymentServer) ChargeCustomer(ctx context.Context, req *paymentv1.ChargeCustomerRequest) (*paymentv1.ChargeCustomerResponse, error) {
	invoiceID, err := uuid.Parse(req.InvoiceId)
	if err != nil {
		return nil, err
	}
	customerID, err := uuid.Parse(req.CustomerId)
	if err != nil {
		return nil, err
	}

	payment, err := s.svc.ChargeCustomer(ctx, invoiceID, customerID, req.AmountCents, req.Currency, req.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	resp := &paymentv1.ChargeCustomerResponse{
		PaymentId:   payment.ID.String(),
		Status:      toProtoStatus(payment.Status),
		ProviderRef: payment.ProviderRef,
		ProcessedAt: timestamppb.New(time.Now()),
	}
	if payment.Status == domain.PaymentFailed {
		resp.FailureReason = "payment declined by gateway"
	}
	return resp, nil
}

func (s *PaymentServer) GetPaymentStatus(ctx context.Context, req *paymentv1.GetPaymentStatusRequest) (*paymentv1.GetPaymentStatusResponse, error) {
	invoiceID, err := uuid.Parse(req.InvoiceId)
	if err != nil {
		return nil, err
	}

	payment, err := s.svc.GetStatus(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	return &paymentv1.GetPaymentStatusResponse{
		PaymentId:    payment.ID.String(),
		Status:       toProtoStatus(payment.Status),
		AttemptCount: int32(payment.AttemptCount),
		UpdatedAt:    timestamppb.New(payment.UpdatedAt),
	}, nil
}

func toProtoStatus(s domain.PaymentStatus) paymentv1.PaymentStatus {
	switch s {
	case domain.PaymentPending:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_PENDING
	case domain.PaymentSucceeded:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_SUCCEEDED
	case domain.PaymentFailed:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_FAILED
	case domain.PaymentRefunded:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_REFUNDED
	default:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}