package domain

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoicePending InvoiceStatus = "pending"
	InvoicePaid    InvoiceStatus = "paid"
	InvoiceFailed  InvoiceStatus = "failed"
	InvoiceVoid    InvoiceStatus = "void"
)

type Invoice struct {
	ID             uuid.UUID     `json:"id"`
	SubscriptionID uuid.UUID     `json:"subscriptionId"`
	CustomerID     uuid.UUID     `json:"customerId"`
	AmountCents    int64         `json:"amountCents"`
	Currency       string        `json:"currency"`
	Status         InvoiceStatus `json:"status"`
	DueDate        time.Time     `json:"dueDate"`
	PaidAt         *time.Time    `json:"paidAt,omitempty"`
	Version        int           `json:"-"`
	CreatedAt      time.Time     `json:"createdAt"`
}