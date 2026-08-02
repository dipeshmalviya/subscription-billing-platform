package repository

import (
	"context"
	"errors"

	"github.com/dipeshmalviya/subscription-billing-platform/payment/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrVersionConflict = errors.New("version conflict")
	ErrNotFound        = errors.New("payment not found")
)

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// GetByIdempotencyKey lets the gRPC handler dedupe: if Subscription Service
// retries a ChargeCustomer call after a timeout, we return the existing
// payment instead of charging twice.
func (r *PaymentRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Payment, error) {
	query := `SELECT id, invoice_id, customer_id, amount_cents, currency, status, provider_ref,
			   idempotency_key, attempt_count, version, created_at, updated_at
			   FROM payments WHERE idempotency_key = $1`
	row := r.db.QueryRow(ctx, query, key)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.InvoiceID, &p.CustomerID, &p.AmountCents, &p.Currency, &p.Status,
		&p.ProviderRef, &p.IdempotencyKey, &p.AttemptCount, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	query := `
		INSERT INTO payments (id, invoice_id, customer_id, amount_cents, currency, status, idempotency_key, attempt_count, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 0, 0, now(), now())`
	_, err := r.db.Exec(ctx, query, p.ID, p.InvoiceID, p.CustomerID, p.AmountCents, p.Currency, p.Status, p.IdempotencyKey)
	return err
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	query := `SELECT id, invoice_id, customer_id, amount_cents, currency, status, provider_ref,
			   idempotency_key, attempt_count, version, created_at, updated_at
			   FROM payments WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.InvoiceID, &p.CustomerID, &p.AmountCents, &p.Currency, &p.Status,
		&p.ProviderRef, &p.IdempotencyKey, &p.AttemptCount, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) GetByInvoiceID(ctx context.Context, invoiceID uuid.UUID) (*domain.Payment, error) {
	query := `SELECT id, invoice_id, customer_id, amount_cents, currency, status, provider_ref,
			   idempotency_key, attempt_count, version, created_at, updated_at
			   FROM payments WHERE invoice_id = $1 ORDER BY created_at DESC LIMIT 1`
	row := r.db.QueryRow(ctx, query, invoiceID)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.InvoiceID, &p.CustomerID, &p.AmountCents, &p.Currency, &p.Status,
		&p.ProviderRef, &p.IdempotencyKey, &p.AttemptCount, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UpdateResult records the outcome of a charge attempt with optimistic locking.
func (r *PaymentRepository) UpdateResult(ctx context.Context, id uuid.UUID, status domain.PaymentStatus, providerRef string, expectedVersion int) error {
	query := `
		UPDATE payments
		SET status = $1, provider_ref = $2, attempt_count = attempt_count + 1,
		    version = version + 1, updated_at = now()
		WHERE id = $3 AND version = $4`
	tag, err := r.db.Exec(ctx, query, status, providerRef, id, expectedVersion)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrVersionConflict
	}
	return nil
}

// ListFailedForRetry finds failed payments under a max attempt threshold —
// used by the retry worker.
func (r *PaymentRepository) ListFailedForRetry(ctx context.Context, maxAttempts, limit int) ([]*domain.Payment, error) {
	query := `SELECT id, invoice_id, customer_id, amount_cents, currency, status, provider_ref,
			   idempotency_key, attempt_count, version, created_at, updated_at
			   FROM payments
			   WHERE status = 'failed' AND attempt_count < $1
			   ORDER BY updated_at ASC LIMIT $2`
	rows, err := r.db.Query(ctx, query, maxAttempts, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.InvoiceID, &p.CustomerID, &p.AmountCents, &p.Currency, &p.Status,
			&p.ProviderRef, &p.IdempotencyKey, &p.AttemptCount, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, &p)
	}
	return payments, rows.Err()
}