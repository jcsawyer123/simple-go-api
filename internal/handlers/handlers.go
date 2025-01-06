package handlers

import (
    "encoding/json"
    "net/http"
    "time"
	
    "github.com/jcsawyer123/simple-go-api/internal/auth"
	"github.com/jcsawyer123/simple-go-api/internal/health"
)

type Handlers struct {
    auth   *auth.Client
    health *health.HealthManager
}

func New(auth *auth.Client, health *health.HealthManager) *Handlers {
    return &Handlers{
        auth: auth,
        health: health,
    }
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
    status := h.health.GetStatus()
    isHealthy := h.health.IsHealthy()
    
    response := map[string]interface{}{
        "status": map[string]interface{}{
            "healthy": isHealthy,
            "checks": status,
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
        "message": "data endpoint",
        "timestamp": time.Now().UTC(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

func (h *Handlers) AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized - No token provided", http.StatusUnauthorized)
            return
        }

        if err := h.auth.ValidateToken(r.Context(), token); err != nil {
            http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}