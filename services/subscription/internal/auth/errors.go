package auth

import "errors"

var (
	ErrForbidden    = errors.New("forbidden: insufficient role")
	ErrUnauthorized = errors.New("unauthorized: missing or invalid token")
)