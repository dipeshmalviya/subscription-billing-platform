package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	CustomerIDKey contextKey = "customerID"
	RoleKey       contextKey = "role"
)

func Middleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				// No token — allow request through unauthenticated (public queries
				// like `plans` don't need auth). Resolvers check context for role.
				next.ServeHTTP(w, r)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtManager.VerifyAccessToken(tokenStr)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), CustomerIDKey, claims.CustomerID)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole is called inside resolvers to enforce admin-only actions.
func RequireRole(ctx context.Context, required string) error {
	role, _ := ctx.Value(RoleKey).(string)
	if role != required {
		return ErrForbidden
	}
	return nil
}

func CustomerIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(CustomerIDKey).(string)
	return id, ok
}