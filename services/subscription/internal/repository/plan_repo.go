package repository

import (
	"context"

	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlanRepository struct {
	db *pgxpool.Pool
}

func NewPlanRepository(db *pgxpool.Pool) *PlanRepository {
	return &PlanRepository{db: db}
}

func (r *PlanRepository) Create(ctx context.Context, p *domain.Plan) error {
	query := `
		INSERT INTO plans (id, name, price_cents, currency, billing_interval, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())`
	_, err := r.db.Exec(ctx, query, p.ID, p.Name, p.PriceCents, p.Currency, p.BillingInterval, p.IsActive)
	return err
}

func (r *PlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Plan, error) {
	query := `SELECT id, name, price_cents, currency, billing_interval, is_active, created_at
			   FROM plans WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	var p domain.Plan
	err := row.Scan(&p.ID, &p.Name, &p.PriceCents, &p.Currency, &p.BillingInterval, &p.IsActive, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PlanRepository) ListActive(ctx context.Context) ([]*domain.Plan, error) {
	query := `SELECT id, name, price_cents, currency, billing_interval, is_active, created_at
			   FROM plans WHERE is_active = true ORDER BY price_cents ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*domain.Plan
	for rows.Next() {
		var p domain.Plan
		if err := rows.Scan(&p.ID, &p.Name, &p.PriceCents, &p.Currency, &p.BillingInterval, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		plans = append(plans, &p)
	}
	return plans, rows.Err()
}