package aims

import (
	"net/http"

	"github.com/jcsawyer123/simple-go-api/internal/auth"
)

// Middleware implements the auth.Middleware interface for AIMS
type Middleware struct {
	service *Client
}

// Ensure Middleware implements the auth.Middleware interface
var _ auth.Middleware = (*Middleware)(nil)

// NewMiddleware creates a new AIMS middleware
func NewMiddleware(service *Client) *Middleware {
	return &Middleware{
		service: service,
	}
}

// Authenticate implements the auth.Middleware interface
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(AimsHeaderName)
		if token == "" {
			http.Error(w, "Unauthorized - No AIMS token provided", http.StatusUnauthorized)
			return
		}

		if err := m.service.ValidateToken(r.Context(), token); err != nil {
			http.Error(w, "Unauthorized - Invalid AIMS token", http.StatusUnauthorized)
			return
		}

		// Store the token in context for later use
		ctx := auth.WithToken(r.Context(), token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermissions implements the auth.Middleware interface
func (m *Middleware) RequirePermissions(requiredPerm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := auth.TokenFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized - No token in context", http.StatusUnauthorized)
				return
			}

			if err := m.service.ValidatePermissions(r.Context(), token, requiredPerm); err != nil {
				http.Error(w, "Forbidden - Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
