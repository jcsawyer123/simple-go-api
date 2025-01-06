package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jcsawyer123/simple-go-api/internal/auth"
	"github.com/jcsawyer123/simple-go-api/internal/health"
)

type ctxKey string

const (
	tokenCtxKey ctxKey = "token"
)

type Handlers struct {
	auth    *auth.Client
	health  *health.HealthManager
	bufPool *sync.Pool // buffer pool for JSON encoding
}

func New(auth *auth.Client, health *health.HealthManager) *Handlers {
	return &Handlers{
		auth:   auth,
		health: health,
		bufPool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

func (h *Handlers) writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	buf := h.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		h.bufPool.Put(buf)
	}()

	if err := json.NewEncoder(buf).Encode(v); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := w.Write(buf.Bytes())
	return err
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

	statusCode := http.StatusOK
	if !isHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	if err := h.writeJSON(w, statusCode, response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) GetData(w http.ResponseWriter, r *http.Request, buf *bytes.Buffer) {
	response := map[string]interface{}{
		"message":   "data endpoint",
		"timestamp": time.Now().UTC(),
	}

	if err := h.writeJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

type TestPermissionsResponse struct {
	HasPermission bool   `json:"has_permission"`
	Message       string `json:"message"`
}

func (h *Handlers) TestPermissions(w http.ResponseWriter, r *http.Request) {
	token, ok := r.Context().Value(tokenCtxKey).(string)
	if !ok {
		http.Error(w, "Internal server error - No token in context", http.StatusInternalServerError)
		return
	}

	err := h.auth.ValidatePermissions(r.Context(), token, "*:managed:*:*")
	response := TestPermissionsResponse{
		HasPermission: err == nil,
		Message:       "Permission check completed",
	}
	if err != nil {
		response.Message = fmt.Sprintf("Permission denied: %v", err)
	}

	if err := h.writeJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("x-aims-auth-token")
		if token == "" {
			http.Error(w, "Unauthorized - No token provided", http.StatusUnauthorized)
			return
		}

		if err := h.auth.ValidateToken(r.Context(), token); err != nil {
			http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
			return
		}

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
