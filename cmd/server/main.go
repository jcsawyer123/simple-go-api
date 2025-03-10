package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jcsawyer123/simple-go-api/internal/config"
	"github.com/jcsawyer123/simple-go-api/internal/logger"
	"github.com/jcsawyer123/simple-go-api/internal/metrics"
	"github.com/jcsawyer123/simple-go-api/internal/profiling"
	"github.com/jcsawyer123/simple-go-api/internal/server"
)

func main() {
	// Load environment variables from .env files
	if err := config.LoadEnv(); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
		// Continue execution - .env file is optional
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Start profiling if enabled
	if cfg.ProfilingPort != "" {
		profiling.Start(cfg.ProfilingPort)
	}

	// setup logger
	logger.Init()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Fatalf("Shutdown signal received")
		cancel()
	}()

	// Clean up metrics on exit
	defer metrics.CloseGlobal()

	svc, err := server.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	logger.Infof("Starting service on port %s", cfg.Port)
	if err := svc.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
