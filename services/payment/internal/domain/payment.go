package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentSucceeded PaymentStatus = "succeeded"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)

type Payment struct {
	ID             uuid.UUID     `json:"id"`
	InvoiceID      uuid.UUID     `json:"invoiceId"`
	CustomerID     uuid.UUID     `json:"customerId"`
	AmountCents    int64         `json:"amountCents"`
	Currency       string        `json:"currency"`
	Status         PaymentStatus `json:"status"`
	ProviderRef    string        `json:"providerRef"`
	IdempotencyKey string        `json:"idempotencyKey"`
	AttemptCount   int           `json:"attemptCount"`
	Version        int           `json:"-"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

type PaymentAttempt struct {
	ID            uuid.UUID
	PaymentID     uuid.UUID
	AttemptNumber int
	Status        PaymentStatus
	ErrorMessage  string
	AttemptedAt   time.Time
}