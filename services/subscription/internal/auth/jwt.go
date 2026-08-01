package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	CustomerID string `json:"sub"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey        string
	refreshSecretKey string
}

func NewJWTManager(secretKey, refreshSecretKey string) *JWTManager {
	return &JWTManager{secretKey: secretKey, refreshSecretKey: refreshSecretKey}
}

func (m *JWTManager) GenerateAccessToken(customerID uuid.UUID, role string) (string, error) {
	claims := Claims{
		CustomerID: customerID.String(),
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) GenerateRefreshToken(customerID uuid.UUID) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   customerID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.refreshSecretKey))
}

func (m *JWTManager) VerifyAccessToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}
	return claims, nil
}