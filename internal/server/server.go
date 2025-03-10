package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jcsawyer123/simple-go-api/internal/auth"
	"github.com/jcsawyer123/simple-go-api/internal/auth/aims"
	"github.com/jcsawyer123/simple-go-api/internal/config"
	"github.com/jcsawyer123/simple-go-api/internal/handlers"
	"github.com/jcsawyer123/simple-go-api/internal/logger"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	httpServer *http.Server
	router     *chi.Mux
	auth       auth.Service
	middleware *Middleware
	handlers   *handlers.Handlers
	bufPool    *sync.Pool
}

func New(cfg *config.Config) (*Server, error) {
	// Create a buffer pool
	bufPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	// Setup Auth Client
	authClient, err := aims.NewClient(cfg.AuthServiceURL)
	if err != nil {
		return nil, fmt.Errorf("creating auth client: %w", err)
	}

	// Initialize the router
	router := chi.NewRouter()

	// Get the auth middleware from the client
	authMiddleware := authClient.CreateMiddleware()

	// Create middleware manager
	middleware := NewMiddleware(authMiddleware)

	// Initialize server
	srv := &Server{
		router:     router,
		auth:       authClient,
		middleware: middleware,
		bufPool:    bufPool,
		handlers:   handlers.New(authClient),
	}
	logger.Info().Msg("Server initialized")

	// Setup middleware and routes
	srv.setupMiddleware()
	srv.setupRoutes()

	logger.Info().Msg("Server started")

	// Setup HTTP server
	srv.httpServer = &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	return srv, nil
}

func (s *Server) setupMiddleware() {
	s.middleware.SetupGlobal(*s.router)
}

func (s *Server) setupRoutes() {
	s.router.Get("/health", s.handlers.HealthCheck)
	s.router.Route("/api", func(r chi.Router) {
		// Auth middleware for all /api routes
		r.Use(s.middleware.Authenticate())

		r.Get("/data", s.handlers.GetData)

		// Permission-protected routes
		r.Route("/perms", func(r chi.Router) {
			r.Use(s.middleware.RequirePermissions(aims.MyServiceUpdatePerm))
			r.Get("/test", s.handlers.TestPermissions)
		})

		// Alternative permission middleware usage - instigator:*:disable:account has explicit deny
		r.With(s.middleware.RequirePermissions(aims.InstigatorDisableAccountPerm)).
			Get("/test", s.handlers.TestPermissions)
	})
}

func (s *Server) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			return fmt.Errorf("http server error: %w", err)
		}
		return nil
	})

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
