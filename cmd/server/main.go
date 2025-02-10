package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jcsawyer123/simple-go-api/internal/config"
	"github.com/jcsawyer123/simple-go-api/internal/logger"
	"github.com/jcsawyer123/simple-go-api/internal/profiling"
	"github.com/jcsawyer123/simple-go-api/internal/server"
)

func main() {
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

	svc, err := server.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	logger.Infof("Starting service on port %s", cfg.Port)
	if err := svc.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
