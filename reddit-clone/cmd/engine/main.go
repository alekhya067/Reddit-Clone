// cmd/engine/main.go
package main

import (
    "flag"
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    
    "reddit-clone/internal/engine"
    "reddit-clone/internal/proto"
    "reddit-clone/internal/server"
    "reddit-clone/pkg/metrics"
)

// Add MetricsServer type
type MetricsServer struct {
    collector *metrics.Collector
}

func main() {
    // Parse configuration
    port := flag.Int("port", 50051, "The server port")
    metricsPort := flag.Int("metrics-port", 50052, "The metrics port")
    metricsInterval := flag.Duration("metrics-interval", time.Minute, "Metrics collection interval")
    flag.Parse()

    // Create components
    redditEngine := engine.NewRedditEngine()
    metricsCollector := metrics.NewCollector()
    redditServer := server.NewRedditServer(redditEngine, metricsCollector)

    // Create gRPC server
    grpcServer := grpc.NewServer()
    proto.RegisterRedditServiceServer(grpcServer, redditServer)
    reflection.Register(grpcServer)

    // Start metrics server
    go startMetricsServer(metricsCollector, *metricsPort)

    // Start listening
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    // Setup metrics collection
    metricsTicker := time.NewTicker(*metricsInterval)
    defer metricsTicker.Stop()

    // Setup graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

    // Start server
    log.Printf("Starting Reddit engine server on port %d\n", *port)
    go func() {
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("failed to serve: %v", err)
        }
    }()

    // Main loop
    for {
        select {
        case <-metricsTicker.C:
            printMetrics(metricsCollector)

        case sig := <-stop:
            log.Printf("Received signal: %v\n", sig)
            log.Println("Gracefully shutting down server...")
            printMetrics(metricsCollector)
            grpcServer.GracefulStop()
            return
        }
    }
}

func startMetricsServer(collector *metrics.Collector, port int) {
    metricsServer := metrics.NewServer(collector)
    addr := fmt.Sprintf(":%d", port)
    log.Printf("Starting metrics server on %s\n", addr)
    if err := metricsServer.ListenAndServe(addr); err != nil {
        log.Printf("Metrics server error: %v\n", err)
    }
}

// Add printMetrics function
func printMetrics(collector *metrics.Collector) {
    stats := collector.GetStats()
    log.Printf("\n=== Server Metrics ===\n")
    log.Printf("Total Requests: %d\n", stats.TotalRequests)
    log.Printf("Average Latency: %v\n", stats.AverageLatency)
    log.Printf("Active Users: %d\n", stats.ActiveUsers)
    log.Printf("Total Posts: %d\n", stats.TotalPosts)
    log.Printf("Total Comments: %d\n", stats.TotalComments)
    log.Printf("Total Votes: %d\n", stats.TotalVotes)
    
    log.Printf("\nSubreddit Statistics:\n")
    for _, stat := range stats.SubredditStats {
        log.Printf("- %s:\n", stat.Name)
        log.Printf("  Members: %d\n", stat.MemberCount)
        log.Printf("  Posts: %d\n", stat.PostCount)
        log.Printf("  Comments: %d\n", stat.CommentCount)
        log.Printf("  Votes: %d\n", stat.VoteCount)
    }
    log.Printf("\n")
}