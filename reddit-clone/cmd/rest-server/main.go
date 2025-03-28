package main

import (
    "flag"
    "log"
    "os"
    "os/signal"
    "syscall"

    "reddit-clone/internal/engine"
    "reddit-clone/internal/rest"
)

func main() {
    // Parse command line arguments
    port := flag.String("port", ":8080", "REST server port")
    enginePort := flag.String("engine-port", ":50051", "gRPC engine port")
    flag.Parse()

    // Create the Reddit engine
    redditEngine := engine.NewRedditEngine()

    // Create and start gRPC server for the engine
    go func() {
        if err := redditEngine.Start(*enginePort); err != nil {
            log.Fatalf("Failed to start engine: %v", err)
        }
    }()

    // Create REST server
    server := rest.NewServer(redditEngine)

    // Setup graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

    // Start REST server in a goroutine
    go func() {
        log.Printf("Starting REST server on port %s", *port)
        if err := server.Start(*port); err != nil {
            log.Fatalf("Failed to start REST server: %v", err)
        }
    }()

    // Wait for interrupt signal
    <-stop
    log.Println("Shutting down server...")
}