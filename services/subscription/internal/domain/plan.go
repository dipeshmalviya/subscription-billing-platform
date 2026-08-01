package domain

import (
	"time"

	"github.com/google/uuid"
)

type BillingInterval string

const (
	IntervalMonthly BillingInterval = "monthly"
	IntervalYearly  BillingInterval = "yearly"
)

type Plan struct {
	ID               uuid.UUID       `json:"id"`
	Name             string          `json:"name"`
	PriceCents       int64           `json:"priceCents"`
	Currency         string          `json:"currency"`
	BillingInterval  BillingInterval `json:"billingInterval"`
	IsActive         bool            `json:"isActive"`
	CreatedAt        time.Time       `json:"createdAt"`
}