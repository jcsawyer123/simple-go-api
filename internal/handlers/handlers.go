package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jcsawyer123/simple-go-api/internal/auth"
	"github.com/jcsawyer123/simple-go-api/internal/health"
)

type ctxKey string

const (
	tokenCtxKey ctxKey = "token"
)

type Handlers struct {
	auth   *auth.Client
	health *health.HealthManager
}

func New(auth *auth.Client, health *health.HealthManager) *Handlers {
	return &Handlers{
		auth:   auth,
		health: health,
	}
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := h.health.GetStatus()
	isHealthy := h.health.IsHealthy()

	response := map[string]interface{}{
		"status": map[string]interface{}{
			"healthy": isHealthy,
			"checks":  status,
		},
		"version": "1.0.0",
	}

	if !isHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) GetData(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message":   "data endpoint",
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type TestPermissionsResponse struct {
	HasPermission bool   `json:"has_permission"`
	Message       string `json:"message"`
}

func (h *Handlers) TestPermissions(w http.ResponseWriter, r *http.Request) {
	// Get the token from the context using the proper context key
	token, ok := r.Context().Value(tokenCtxKey).(string)
	// log r.Context() possible values
	print(r.Context())

	if !ok {
		http.Error(w, "Internal server error - No token in context", http.StatusInternalServerError)
		return
	}

	// Test for the specific permission
	err := h.auth.ValidatePermissions(r.Context(), token, "*:managed:*:*")

	response := TestPermissionsResponse{
		HasPermission: err == nil,
		Message:       "Permission check completed",
	}

	if err != nil {
		response.Message = fmt.Sprintf("Permission denied: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from x-aims-auth-token header
		token := r.Header.Get("x-aims-auth-token")
		if token == "" {
			http.Error(w, "Unauthorized - No token provided", http.StatusUnauthorized)
			return
		}

		// Validate the token
		if err := h.auth.ValidateToken(r.Context(), token); err != nil {
			http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
			return
		}

		// Set token in context using the defined context key
		ctx := context.WithValue(r.Context(), tokenCtxKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handlers) RequirePermissionsMiddleware(requiredPerm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := r.Context().Value(tokenCtxKey).(string)
			if !ok {
				http.Error(w, "Unauthorized - No token in context", http.StatusUnauthorized)
				return
			}

			if err := h.auth.ValidatePermissions(r.Context(), token, requiredPerm); err != nil {
				http.Error(w, "Forbidden - Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
