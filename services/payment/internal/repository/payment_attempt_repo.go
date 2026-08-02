package repository

import (
	"context"

	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentAttemptRepository struct {
	db *pgxpool.Pool
}

func NewPaymentAttemptRepository(db *pgxpool.Pool) *PaymentAttemptRepository {
	return &PaymentAttemptRepository{db: db}
}

func (r *PaymentAttemptRepository) Record(ctx context.Context, a *domain.PaymentAttempt) error {
	query := `
		INSERT INTO payment_attempts (id, payment_id, attempt_number, status, error_message, attempted_at)
		VALUES ($1, $2, $3, $4, $5, now())`
	_, err := r.db.Exec(ctx, query, a.ID, a.PaymentID, a.AttemptNumber, a.Status, a.ErrorMessage)
	return err
}