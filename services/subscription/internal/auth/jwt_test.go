package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerateAndVerifyAccessToken(t *testing.T) {
	manager := NewJWTManager("secret-key", "refresh-secret-key")
	customerID := uuid.New()
	token, err := manager.GenerateAccessToken(customerID, "admin")
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	claims, err := manager.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("VerifyAccessToken failed: %v", err)
	}

	if claims.CustomerID != customerID.String() {
		t.Fatalf("expected customer ID %q, got %q", customerID.String(), claims.CustomerID)
	}
	if claims.Role != "admin" {
		t.Fatalf("expected role admin, got %q", claims.Role)
	}
}

func TestVerifyAccessToken_InvalidToken(t *testing.T) {
	manager := NewJWTManager("secret-key", "refresh-secret-key")
	_, err := manager.VerifyAccessToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
