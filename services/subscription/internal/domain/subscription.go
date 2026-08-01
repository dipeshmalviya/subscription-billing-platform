package domain

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionStatus string

const (
	StatusActive   SubscriptionStatus = "active"
	StatusCanceled SubscriptionStatus = "canceled"
	StatusPastDue  SubscriptionStatus = "past_due"
	StatusPaused   SubscriptionStatus = "paused"
)

type Subscription struct {
	ID                  uuid.UUID          `json:"id"`
	CustomerID          uuid.UUID          `json:"customerId"`
	PlanID              uuid.UUID          `json:"planId"`
	Status              SubscriptionStatus `json:"status"`
	CurrentPeriodStart  time.Time          `json:"currentPeriodStart"`
	CurrentPeriodEnd    time.Time          `json:"currentPeriodEnd"`
	CanceledAt          *time.Time         `json:"canceledAt,omitempty"`
	Version             int                `json:"-"`
	CreatedAt           time.Time          `json:"createdAt"`
	UpdatedAt           time.Time          `json:"updatedAt"`
}