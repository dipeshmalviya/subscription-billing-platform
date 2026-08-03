package gateway

import (
	"context"
	"testing"
)

func TestMockGateway_Refund(t *testing.T) {
	g := NewMockGateway()
	if err := g.Refund(context.Background(), "provider-abc"); err != nil {
		t.Fatalf("Refund returned error: %v", err)
	}
}

func TestMockGateway_Charge(t *testing.T) {
	g := NewMockGateway()
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		result, err := g.Charge(ctx, 1000, "USD")
		if err != nil && err != ErrGatewayDeclined {
			t.Fatalf("Charge returned unexpected error: %v", err)
		}
		if err == nil {
			if result == nil {
				t.Fatal("expected non-nil result when charge succeeds")
			}
			if result.ProviderRef == "" {
				t.Fatal("expected provider ref to be set")
			}
		}
	}
}
