package repository

import (
	"context"
	"time"

	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerRepository struct {
	db *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Create(ctx context.Context, c *domain.Customer, passwordHash string) error {
	query := `
		INSERT INTO customers (id, email, full_name, role, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())`
	_, err := r.db.Exec(ctx, query, c.ID, c.Email, c.FullName, c.Role, passwordHash)
	return err
}

func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	query := `SELECT id, email, full_name, role, created_at, updated_at FROM customers WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	var c domain.Customer
	err := row.Scan(&c.ID, &c.Email, &c.FullName, &c.Role, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepository) GetByEmail(ctx context.Context, email string) (*domain.Customer, string, error) {
	query := `SELECT id, email, full_name, role, password_hash, created_at, updated_at FROM customers WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)

	var c domain.Customer
	var passwordHash string
	err := row.Scan(&c.ID, &c.Email, &c.FullName, &c.Role, &passwordHash, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, "", err
	}
	return &c, passwordHash, nil
}

func (r *CustomerRepository) List(ctx context.Context, limit, offset int) ([]*domain.Customer, error) {
	query := `SELECT id, email, full_name, role, created_at, updated_at
			   FROM customers ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []*domain.Customer
	for rows.Next() {
		var c domain.Customer
		if err := rows.Scan(&c.ID, &c.Email, &c.FullName, &c.Role, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		customers = append(customers, &c)
	}
	return customers, rows.Err()
}

var _ = time.Now // placeholder import guard if unused elsewhere