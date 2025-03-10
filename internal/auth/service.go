// internal/auth/service.go
package auth

import (
	"context"
	"net/http"
)

// Service defines the generic interface for authentication and authorization
type Service interface {
	// ValidateToken validates a token against the auth service
	ValidateToken(ctx context.Context, token string) error

	// ValidatePermissions checks if the token has the required permission
	// The permission format and validation logic is implementation-specific
	ValidatePermissions(ctx context.Context, token, requiredPerm string) error

	// CreateMiddleware returns a middleware for this auth service
	CreateMiddleware() Middleware
}

// Middleware provides HTTP middleware for authentication and authorization
type Middleware interface {
	// Authenticate is a middleware that verifies the authentication token
	Authenticate(next http.Handler) http.Handler

	// RequirePermissions is a middleware factory that creates middleware requiring specific permissions
	RequirePermissions(requiredPerm string) func(http.Handler) http.Handler
}
