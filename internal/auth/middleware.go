package auth

import (
	"context"
	"net/http"
)

type Middleware struct {
	client *AuthClient
}

func NewMiddleware(client *AuthClient) *Middleware {
	return &Middleware{
		client: client,
	}
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("x-aims-auth-token")
		if token == "" {
			http.Error(w, "Unauthorized - No token provided", http.StatusUnauthorized)
			return
		}

		if err := m.client.ValidateToken(r.Context(), token); err != nil {
			http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), TokenCtxKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) RequirePermissions(requiredPerm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := r.Context().Value(TokenCtxKey).(string)
			if !ok {
				http.Error(w, "Unauthorized - No token in context", http.StatusUnauthorized)
				return
			}

			if err := m.client.ValidatePermissions(r.Context(), token, requiredPerm); err != nil {
				http.Error(w, "Forbidden - Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
