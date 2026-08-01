package repository

import (
	"context"

	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvoiceRepository struct {
	db *pgxpool.Pool
}

func NewInvoiceRepository(db *pgxpool.Pool) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Create(ctx context.Context, inv *domain.Invoice) error {
	query := `
		INSERT INTO invoices (id, subscription_id, customer_id, amount_cents, currency, status, due_date, version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 0, now())`
	_, err := r.db.Exec(ctx, query, inv.ID, inv.SubscriptionID, inv.CustomerID, inv.AmountCents, inv.Currency, inv.Status, inv.DueDate)
	return err
}

func (r *InvoiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	query := `SELECT id, subscription_id, customer_id, amount_cents, currency, status, due_date, paid_at, version, created_at
			   FROM invoices WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	var inv domain.Invoice
	err := row.Scan(&inv.ID, &inv.SubscriptionID, &inv.CustomerID, &inv.AmountCents, &inv.Currency,
		&inv.Status, &inv.DueDate, &inv.PaidAt, &inv.Version, &inv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *InvoiceRepository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*domain.Invoice, error) {
	query := `SELECT id, subscription_id, customer_id, amount_cents, currency, status, due_date, paid_at, version, created_at
			   FROM invoices WHERE customer_id = $1 ORDER BY due_date DESC`
	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		var inv domain.Invoice
		if err := rows.Scan(&inv.ID, &inv.SubscriptionID, &inv.CustomerID, &inv.AmountCents, &inv.Currency,
			&inv.Status, &inv.DueDate, &inv.PaidAt, &inv.Version, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, &inv)
	}
	return invoices, rows.Err()
}

func (r *InvoiceRepository) MarkPaid(ctx context.Context, id uuid.UUID, expectedVersion int) error {
	query := `
		UPDATE invoices
		SET status = 'paid', paid_at = now(), version = version + 1
		WHERE id = $1 AND version = $2`
	tag, err := r.db.Exec(ctx, query, id, expectedVersion)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrVersionConflict
	}
	return nil
}