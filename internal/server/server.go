package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
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
	bufPool    *sync.Pool
}

func New(cfg *config.Config) (*Server, error) {

	// Create a buffer pool
	bufPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	// Initialize the auth client
	authClient, err := auth.NewClient(cfg.AuthServiceURL)
	if err != nil {
		return nil, fmt.Errorf("creating auth client: %w", err)
	}

	// Initialize the health manager
	healthManager := health.NewHealthManager(5 * time.Minute)
	healthManager.RegisterChecker("auth", authClient)

	// Initialize the router
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Logger)

	h := handlers.New(authClient, healthManager)

	r.Get("/health", h.HealthCheck)
	r.Route("/api", func(r chi.Router) {
		r.Use(h.AuthMiddleware)
		r.Get("/data", func(w http.ResponseWriter, r *http.Request) {
			buf := bufPool.Get().(*bytes.Buffer)
			defer bufPool.Put(buf)
			buf.Reset()

			h.GetData(w, r, buf)
		})

		r.Route("/perms", func(r chi.Router) {
			r.Use(h.RequirePermissionsMiddleware("*:managed:*:*"))
			r.Get("/test", h.TestPermissions)
		})

		r.With(h.RequirePermissionsMiddleware("*:managed:*:*")).
			Get("/test", h.TestPermissions)
	})

	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           r,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       120 * time.Second,
			ReadHeaderTimeout: 2 * time.Second,
			MaxHeaderBytes:    1 << 20,
		},
		auth:    authClient,
		health:  healthManager,
		bufPool: bufPool,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	s.health.StartChecks(ctx)

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
