package gateway

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

var ErrGatewayDeclined = errors.New("payment declined by gateway")

type ChargeResult struct {
	ProviderRef string
	Success     bool
}

type MockGateway struct{}

func NewMockGateway() *MockGateway {
	return &MockGateway{}
}

// Charge simulates calling an external payment processor (Stripe-like).
// ~15% failure rate to exercise the retry worker realistically.
func (g *MockGateway) Charge(ctx context.Context, amountCents int64, currency string) (*ChargeResult, error) {
	time.Sleep(50 * time.Millisecond) // simulate network latency

	if rand.Intn(100) < 15 {
		return nil, ErrGatewayDeclined
	}

	return &ChargeResult{
		ProviderRef: fmt.Sprintf("ch_%s", uuid.New().String()[:12]),
		Success:     true,
	}, nil
}

func (g *MockGateway) Refund(ctx context.Context, providerRef string) error {
	time.Sleep(50 * time.Millisecond)
	return nil // mock always succeeds
}