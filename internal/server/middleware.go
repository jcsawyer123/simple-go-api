package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jcsawyer123/simple-go-api/internal/auth"
)

type Middleware struct {
	auth auth.Middleware
}

// NewMiddleware creates a server middleware with the provided auth middleware
func NewMiddleware(authMiddleware auth.Middleware) *Middleware {
	return &Middleware{
		auth: authMiddleware,
	}
}

// SetupGlobal sets up global middleware for the router
func (m *Middleware) SetupGlobal(r chi.Mux) {
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Logger)
}

// Helper methods to access specific middleware
func (m *Middleware) Authenticate() func(http.Handler) http.Handler {
	return m.auth.Authenticate
}

func (m *Middleware) RequirePermissions(perm string) func(http.Handler) http.Handler {
	return m.auth.RequirePermissions(perm)
}
