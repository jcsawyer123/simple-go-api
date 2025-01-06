package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"

	"github.com/jcsawyer123/simple-go-api/internal/auth"
	"github.com/jcsawyer123/simple-go-api/internal/config"
	"github.com/jcsawyer123/simple-go-api/internal/handlers"
	"github.com/jcsawyer123/simple-go-api/internal/health"
)

type Server struct {
	httpServer *http.Server
	auth       *auth.Client
	health     *health.HealthManager
}

func New(cfg *config.Config) (*Server, error) {
	authClient, err := auth.NewClient(cfg.AuthServiceURL)
	if err != nil {
		return nil, fmt.Errorf("creating auth client: %w", err)
	}

	// Create health manager with 5-minute check interval
	healthManager := health.NewHealthManager(5 * time.Minute)

	// Register health checks
	healthManager.RegisterChecker("auth", authClient)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	h := handlers.New(authClient, healthManager)

	// Routes
	r.Get("/health", h.HealthCheck)
	r.Route("/api", func(r chi.Router) {
		r.Use(h.AuthMiddleware)
		r.Get("/data", h.GetData)

		// Routes requiring specific permissions
		r.Route("/perms", func(r chi.Router) {
			r.Use(h.RequirePermissionsMiddleware("*:test:*:*"))
			r.Get("/test", h.TestPermissions)
		})

		// Route without additional permission requirements
		r.With(h.RequirePermissionsMiddleware("*:managed:*:*")).Get("/test", h.TestPermissions)

	})

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: r,
		},
		auth:   authClient,
		health: healthManager,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	// Start health checks
	s.health.StartChecks(ctx)

	// Start HTTP server
	g.Go(func() error {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			return fmt.Errorf("http server error: %w", err)
		}
		return nil
	})

	// Handle graceful shutdown
	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http server shutdown: %w", err)
		}
		return nil
	})

	return g.Wait()
}
