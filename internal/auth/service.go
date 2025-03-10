package auth

import (
	"context"
)

// Service defines the generic interface for authentication and authorization
type Service interface {
	// ValidateToken validates a token against the auth service
	ValidateToken(ctx context.Context, token string) error

	// ValidatePermissions checks if the token has the required permission
	// The permission format and validation logic is implementation-specific
	ValidatePermissions(ctx context.Context, token, requiredPerm string) error
}
