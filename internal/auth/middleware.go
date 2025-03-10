package auth

import (
	"net/http"
)

// Middleware provides HTTP middleware for authentication and authorization
type Middleware interface {
	// Authenticate is a middleware that verifies the authentication token
	Authenticate(next http.Handler) http.Handler

	// RequirePermissions is a middleware factory that creates middleware requiring specific permissions
	RequirePermissions(requiredPerm string) func(http.Handler) http.Handler
}
