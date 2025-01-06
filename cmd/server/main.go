package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/jcsawyer123/simple-go-api/internal/server"
    "github.com/jcsawyer123/simple-go-api/internal/config"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle shutdown gracefully
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        log.Println("Shutdown signal received")
        cancel()
    }()

    svc, err := server.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Starting service on port %s", cfg.Port)
    if err := svc.Start(ctx); err != nil {
        log.Fatal(err)
    }
}