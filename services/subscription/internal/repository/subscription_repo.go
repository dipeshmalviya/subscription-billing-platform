package repository

import (
	"context"
	"errors"
	"time"

	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrVersionConflict = errors.New("version conflict: subscription was modified by another process")

type SubscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, s *domain.Subscription) error {
	query := `
		INSERT INTO subscriptions (id, customer_id, plan_id, status, current_period_start, current_period_end, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 0, now(), now())`
	_, err := r.db.Exec(ctx, query, s.ID, s.CustomerID, s.PlanID, s.Status, s.CurrentPeriodStart, s.CurrentPeriodEnd)
	return err
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	query := `SELECT id, customer_id, plan_id, status, current_period_start, current_period_end,
			   canceled_at, version, created_at, updated_at FROM subscriptions WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	var s domain.Subscription
	err := row.Scan(&s.ID, &s.CustomerID, &s.PlanID, &s.Status, &s.CurrentPeriodStart,
		&s.CurrentPeriodEnd, &s.CanceledAt, &s.Version, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubscriptionRepository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*domain.Subscription, error) {
	query := `SELECT id, customer_id, plan_id, status, current_period_start, current_period_end,
			   canceled_at, version, created_at, updated_at
			   FROM subscriptions WHERE customer_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*domain.Subscription
	for rows.Next() {
		var s domain.Subscription
		if err := rows.Scan(&s.ID, &s.CustomerID, &s.PlanID, &s.Status, &s.CurrentPeriodStart,
			&s.CurrentPeriodEnd, &s.CanceledAt, &s.Version, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, &s)
	}
	return subs, rows.Err()
}

// DueForRenewal finds active subscriptions whose current period has ended.
// Used by the renewal worker — relies on the composite index on (status, current_period_end).
func (r *SubscriptionRepository) DueForRenewal(ctx context.Context, asOf time.Time, limit int) ([]*domain.Subscription, error) {
	query := `SELECT id, customer_id, plan_id, status, current_period_start, current_period_end,
			   canceled_at, version, created_at, updated_at
			   FROM subscriptions
			   WHERE status = 'active' AND current_period_end <= $1
			   ORDER BY current_period_end ASC LIMIT $2`
	rows, err := r.db.Query(ctx, query, asOf, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*domain.Subscription
	for rows.Next() {
		var s domain.Subscription
		if err := rows.Scan(&s.ID, &s.CustomerID, &s.PlanID, &s.Status, &s.CurrentPeriodStart,
			&s.CurrentPeriodEnd, &s.CanceledAt, &s.Version, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, &s)
	}
	return subs, rows.Err()
}

// UpdateStatus performs an optimistic-locked status change.
// If the row's version doesn't match expectedVersion, ErrVersionConflict is returned
// so the caller can reload and retry — this is what protects against two workers
// (or a worker + a user cancel request) racing on the same subscription.
func (r *SubscriptionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus domain.SubscriptionStatus, expectedVersion int) error {
	query := `
		UPDATE subscriptions
		SET status = $1, version = version + 1, updated_at = now(),
		    canceled_at = CASE WHEN $1 = 'canceled' THEN now() ELSE canceled_at END
		WHERE id = $2 AND version = $3`
	tag, err := r.db.Exec(ctx, query, newStatus, id, expectedVersion)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrVersionConflict
	}
	return nil
}

// RenewPeriod advances the billing period after a successful renewal charge.
func (r *SubscriptionRepository) RenewPeriod(ctx context.Context, id uuid.UUID, newStart, newEnd time.Time, expectedVersion int) error {
	query := `
		UPDATE subscriptions
		SET current_period_start = $1, current_period_end = $2, version = version + 1, updated_at = now()
		WHERE id = $3 AND version = $4`
	tag, err := r.db.Exec(ctx, query, newStart, newEnd, id, expectedVersion)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrVersionConflict
	}
	return nil
}