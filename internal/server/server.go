package server

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jcsawyer123/simple-go-api/internal/auth"
	"github.com/jcsawyer123/simple-go-api/internal/auth/aims"
	"github.com/jcsawyer123/simple-go-api/internal/config"
	"github.com/jcsawyer123/simple-go-api/internal/handlers"
	"github.com/jcsawyer123/simple-go-api/internal/logger"
	"github.com/jcsawyer123/simple-go-api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	httpServer     *http.Server
	router         *chi.Mux
	auth           auth.Service
	middleware     *Middleware
	handlers       *handlers.Handlers
	bufPool        *sync.Pool
	metricsHandler http.Handler
}

func New(cfg *config.Config) (*Server, error) {
	// Create a buffer pool
	bufPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	// Initialize metrics system
	var metricsHandler http.Handler
	if err := setupMetrics(cfg.Metrics); err != nil {
		return nil, fmt.Errorf("setting up metrics: %w", err)
	}

	// Get the Prometheus metrics handler if enabled
	if cfg.Metrics.Prometheus.Enabled {
		metricsHandler = promhttp.Handler()
	}

	// Setup Auth Client
	authClient, err := aims.NewClient(cfg.AuthServiceURL)
	if err != nil {
		return nil, fmt.Errorf("creating auth client: %w", err)
	}
	// Get the auth middleware
	authMiddleware, err := auth.NewMiddleware(authClient)
	if err != nil {
		log.Fatalf("Failed to create auth middleware: %v", err)
	}

	// Initialize the router
	router := chi.NewRouter()

	// Create middleware manager
	middleware := NewMiddleware(authMiddleware)

	// Initialize server
	srv := &Server{
		router:         router,
		auth:           authClient,
		middleware:     middleware,
		bufPool:        bufPool,
		handlers:       handlers.New(authClient),
		metricsHandler: metricsHandler,
	}
	logger.Info().Msg("Server initialized")

	// Setup middleware and routes
	srv.setupMiddleware(cfg)
	srv.setupRoutes(cfg)

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

func setupMetrics(cfg config.MetricsConfig) error {
	if !cfg.Enabled {
		// Use a null provider if metrics are disabled
		metrics.InitGlobal(metrics.NewNullProvider())
		return nil
	}

	var providers []metrics.MetricsProvider

	// Setup Prometheus if enabled
	if cfg.Prometheus.Enabled {
		promProvider := metrics.NewPrometheusProvider(metrics.PrometheusConfig{
			Namespace: cfg.Prometheus.Namespace,
			Subsystem: cfg.Prometheus.Subsystem,
		})

		providers = append(providers, promProvider)
		logger.Info().Msgf("Prometheus metrics enabled on the main server at /metrics")
	}

	// // Setup Datadog if enabled
	// if cfg.Datadog.Enabled {
	// 	ddProvider, err := metrics.NewDatadogProvider(metrics.DatadogConfig{
	// 		Address:     cfg.Datadog.Address,
	// 		Namespace:   cfg.Datadog.Namespace,
	// 		DefaultTags: cfg.Datadog.DefaultTags,
	// 	})
	// 	if err != nil {
	// 		return fmt.Errorf("creating Datadog metrics provider: %w", err)
	// 	}

	// 	providers = append(providers, ddProvider)
	// 	logger.Info().Msgf("Datadog metrics enabled with statsd at %s", cfg.Datadog.Address)
	// }

	// Initialize the global metrics reporter with all enabled providers
	if err := metrics.InitGlobal(providers...); err != nil {
		return fmt.Errorf("initializing metrics providers: %w", err)
	}

	return nil
}

func (s *Server) setupMiddleware(cfg *config.Config) {
	s.middleware.SetupGlobal(*s.router)

	// Add metrics middleware if metrics are enabled
	if cfg.Metrics.Enabled {
		s.router.Use(metrics.HTTPMiddleware("http"))
	}
}

func (s *Server) setupRoutes(cfg *config.Config) {
	// Add Prometheus metrics endpoint if enabled
	if cfg.Metrics.Enabled && cfg.Metrics.Prometheus.Enabled && s.metricsHandler != nil {
		s.router.Handle("/metrics", s.metricsHandler)
		logger.Info().Msg("Prometheus metrics endpoint exposed at /metrics")
	}

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

		// Close metrics system on shutdown
		defer metrics.CloseGlobal()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http server shutdown: %w", err)
		}
		return nil
	})

	return g.Wait()
}
