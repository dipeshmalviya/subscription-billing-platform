package service

import (
	"context"
	"errors"

	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/auth"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/domain"
	"github.com/dipeshmalviya/subscription-billing-platform/subscription/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type AuthService struct {
	customerRepo *repository.CustomerRepository
	jwtManager   *auth.JWTManager
}

func NewAuthService(customerRepo *repository.CustomerRepository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{customerRepo: customerRepo, jwtManager: jwtManager}
}

func (s *AuthService) Signup(ctx context.Context, email, fullName, password string) (*domain.Customer, string, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", err
	}

	customer := &domain.Customer{
		ID:       uuid.New(),
		Email:    email,
		FullName: fullName,
		Role:     domain.RoleCustomer,
	}

	if err := s.customerRepo.Create(ctx, customer, string(hash)); err != nil {
		return nil, "", "", err
	}

	access, err := s.jwtManager.GenerateAccessToken(customer.ID, string(customer.Role))
	if err != nil {
		return nil, "", "", err
	}
	refresh, err := s.jwtManager.GenerateRefreshToken(customer.ID)
	if err != nil {
		return nil, "", "", err
	}

	return customer, access, refresh, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.Customer, string, string, error) {
	customer, hash, err := s.customerRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	access, err := s.jwtManager.GenerateAccessToken(customer.ID, string(customer.Role))
	if err != nil {
		return nil, "", "", err
	}
	refresh, err := s.jwtManager.GenerateRefreshToken(customer.ID)
	if err != nil {
		return nil, "", "", err
	}

	return customer, access, refresh, nil
}