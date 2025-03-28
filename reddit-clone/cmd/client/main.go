// cmd/client/main.go
package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "reddit-clone/internal/client"
    "reddit-clone/internal/simulator"
    "reddit-clone/pkg/metrics"
)

type Config struct {
    ServerAddr      string
    NumUsers        int
    Duration        time.Duration
    MetricsInterval time.Duration
    MetricsPort     int
}

func main() {
    // Parse configuration
    config := Config{}
    flag.StringVar(&config.ServerAddr, "server", "localhost:50051", "The server address")
    flag.IntVar(&config.NumUsers, "users", 1000, "Number of users to simulate")
    flag.DurationVar(&config.Duration, "duration", 10*time.Minute, "Duration to run the simulation")
    flag.DurationVar(&config.MetricsInterval, "metrics-interval", time.Minute, "Interval for metrics collection")
    flag.IntVar(&config.MetricsPort, "metrics-port", 50053, "Port for metrics server")
    flag.Parse()

    // Create Reddit client
    redditClient, err := client.NewRedditClient(config.ServerAddr)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer redditClient.Close()

    // Create simulator
    sim := simulator.NewSimulator(redditClient, config.NumUsers)
    metricsCollector := metrics.NewCollector()

    // Setup metrics server
    go startMetricsServer(metricsCollector, config.MetricsPort)

    // Setup metrics collection
    metricsTicker := time.NewTicker(config.MetricsInterval)
    defer metricsTicker.Stop()

    // Setup simulation end timer
    simTimer := time.NewTimer(config.Duration)
    defer simTimer.Stop()

    // Setup graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

    // Start simulation
    log.Printf("Starting simulation with %d users\n", config.NumUsers)
    sim.Start()

    // Main loop
    for {
        select {
        case <-metricsTicker.C:
            metrics := sim.GetMetrics()
            metricsCollector.Update(metrics)
            logMetrics(metricsCollector.GetStats())

        case <-simTimer.C:
            log.Println("Simulation duration completed")
            metrics := sim.GetMetrics()
            metricsCollector.Update(metrics)
            logMetrics(metricsCollector.GetStats())
            sim.Stop()
            return

        case sig := <-stop:
            log.Printf("Received signal: %v\n", sig)
            log.Println("Stopping simulation...")
            metrics := sim.GetMetrics()
            metricsCollector.Update(metrics)
            logMetrics(metricsCollector.GetStats())
            sim.Stop()
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

func logMetrics(stats *metrics.Stats) {
    log.Printf("\n=== Simulation Metrics ===\n")
    log.Printf("Active Users: %d/%d (%.1f%%)\n",
        stats.ActiveUsers,
        stats.TotalUsers,
        float64(stats.ActiveUsers)/float64(stats.TotalUsers)*100)
    
    log.Printf("Average Response Time: %v\n", stats.AverageLatency)
    log.Printf("Total Requests: %d\n", stats.TotalRequests)
    log.Printf("Request Rate: %.2f/sec\n", stats.RequestRate)
    
    log.Printf("\nContent Statistics:\n")
    log.Printf("Total Posts: %d\n", stats.TotalPosts)
    log.Printf("Total Comments: %d\n", stats.TotalComments)
    log.Printf("Total Votes: %d\n", stats.TotalVotes)
    
    log.Printf("\nSubreddit Activity:\n")
    for _, stat := range stats.SubredditStats {
        log.Printf("- %s:\n", stat.Name)
        log.Printf("  Members: %d\n", stat.MemberCount)
        log.Printf("  Posts: %d (%.2f per member)\n",
            stat.PostCount,
            float64(stat.PostCount)/float64(stat.MemberCount))
        log.Printf("  Comments: %d\n", stat.CommentCount)
        log.Printf("  Votes: %d\n", stat.VoteCount)
    }
    log.Printf("\n")
}