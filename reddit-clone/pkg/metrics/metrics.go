package metrics

import (
    "encoding/json"
    "fmt"  // Add this import
    "net/http"
    "sync"
    "time"
    
    "reddit-clone/internal/models"
)

// Stats represents the collected metrics
type Stats struct {
    StartTime       time.Time
    TotalUsers      int64
    ActiveUsers     int64
    TotalPosts     int64
    TotalComments  int64
    TotalVotes     int64
    TotalRequests  int64
    ErrorCount     int64
    RequestRate    float64
    AverageLatency time.Duration
    EndpointStats  map[string]*EndpointStats  // Stats per endpoint
    SubredditStats map[string]*SubredditStats // Stats per subreddit
}

// EndpointStats tracks metrics for each gRPC endpoint
type EndpointStats struct {
    Method         string
    CallCount      int64
    ErrorCount     int64
    TotalLatency   time.Duration
    AverageLatency time.Duration
    LastCall       time.Time
}

// SubredditStats tracks metrics for each subreddit
type SubredditStats struct {
    Name          string
    MemberCount   int64
    PostCount     int64
    CommentCount  int64
    VoteCount     int64
    ActiveUsers   int64
    PopularPosts  []string // IDs of most upvoted posts
}

// Collector manages metrics collection
type Collector struct {
    mtx           sync.RWMutex
    stats         *Stats
    latencies     []time.Duration
    lastUpdate    time.Time
    requestCounts map[string]int64 // requests per second tracking
}

func NewCollector() *Collector {
    return &Collector{
        stats: &Stats{
            StartTime:      time.Now(),
            EndpointStats:  make(map[string]*EndpointStats),
            SubredditStats: make(map[string]*SubredditStats),
        },
        latencies:     make([]time.Duration, 0),
        requestCounts: make(map[string]int64),
    }
}

// RecordLatency records the latency for a specific endpoint
func (c *Collector) RecordLatency(endpoint string, duration time.Duration) {
    c.mtx.Lock()
    defer c.mtx.Unlock()

    stats, exists := c.stats.EndpointStats[endpoint]
    if !exists {
        stats = &EndpointStats{Method: endpoint}
        c.stats.EndpointStats[endpoint] = stats
    }

    stats.CallCount++
    stats.TotalLatency += duration
    stats.AverageLatency = stats.TotalLatency / time.Duration(stats.CallCount)
    stats.LastCall = time.Now()

    c.latencies = append(c.latencies, duration)
    c.updateAverageLatency()
}

// RecordError records an error for a specific endpoint
func (c *Collector) RecordError(endpoint string) {
    c.mtx.Lock()
    defer c.mtx.Unlock()

    stats, exists := c.stats.EndpointStats[endpoint]
    if !exists {
        stats = &EndpointStats{Method: endpoint}
        c.stats.EndpointStats[endpoint] = stats
    }

    stats.ErrorCount++
    c.stats.ErrorCount++
}

// Update updates the overall metrics
func (c *Collector) Update(metrics *models.Metrics) {
    c.mtx.Lock()
    defer c.mtx.Unlock()

    c.stats.TotalUsers = metrics.TotalUsers
    c.stats.ActiveUsers = metrics.ActiveUsers
    c.stats.TotalPosts = metrics.TotalPosts
    c.stats.TotalComments = metrics.TotalComments
    c.stats.TotalVotes = metrics.TotalVotes

    // Update subreddit stats
    for id, stats := range metrics.SubredditStats {
        if _, exists := c.stats.SubredditStats[id]; !exists {
            c.stats.SubredditStats[id] = &SubredditStats{
                Name: stats.Name,
            }
        }
        
        subredditStats := c.stats.SubredditStats[id]
        subredditStats.MemberCount = stats.MemberCount
        subredditStats.PostCount = stats.PostCount
        subredditStats.CommentCount = stats.CommentCount
        subredditStats.VoteCount = stats.VoteCount
        subredditStats.ActiveUsers = stats.ActiveUsers
    }

    // Calculate request rate
    now := time.Now()
    if !c.lastUpdate.IsZero() {
        duration := now.Sub(c.lastUpdate).Seconds()
        if duration > 0 {
            totalRequests := int64(0)
            for _, stats := range c.stats.EndpointStats {
                totalRequests += stats.CallCount
            }
            c.stats.RequestRate = float64(totalRequests-c.stats.TotalRequests) / duration
            c.stats.TotalRequests = totalRequests
        }
    }
    c.lastUpdate = now
}

func (c *Collector) updateAverageLatency() {
    if len(c.latencies) == 0 {
        return
    }

    var total time.Duration
    for _, latency := range c.latencies {
        total += latency
    }
    c.stats.AverageLatency = total / time.Duration(len(c.latencies))
}

// GetStats returns a copy of the current stats
func (c *Collector) GetStats() *Stats {
    c.mtx.RLock()
    defer c.mtx.RUnlock()

    // Create a deep copy of stats
    statsCopy := &Stats{
        StartTime:      c.stats.StartTime,
        TotalUsers:     c.stats.TotalUsers,
        ActiveUsers:    c.stats.ActiveUsers,
        TotalPosts:    c.stats.TotalPosts,
        TotalComments: c.stats.TotalComments,
        TotalVotes:    c.stats.TotalVotes,
        TotalRequests: c.stats.TotalRequests,
        ErrorCount:    c.stats.ErrorCount,
        RequestRate:   c.stats.RequestRate,
        AverageLatency: c.stats.AverageLatency,
        EndpointStats:  make(map[string]*EndpointStats),
        SubredditStats: make(map[string]*SubredditStats),
    }

    // Copy endpoint stats
    for k, v := range c.stats.EndpointStats {
        statsCopy.EndpointStats[k] = &EndpointStats{
            Method:         v.Method,
            CallCount:      v.CallCount,
            ErrorCount:     v.ErrorCount,
            TotalLatency:   v.TotalLatency,
            AverageLatency: v.AverageLatency,
            LastCall:       v.LastCall,
        }
    }

    // Copy subreddit stats
    for k, v := range c.stats.SubredditStats {
        statsCopy.SubredditStats[k] = &SubredditStats{
            Name:         v.Name,
            MemberCount:  v.MemberCount,
            PostCount:    v.PostCount,
            CommentCount: v.CommentCount,
            VoteCount:    v.VoteCount,
            ActiveUsers:  v.ActiveUsers,
            PopularPosts: append([]string{}, v.PopularPosts...),
        }
    }

    return statsCopy
}

// MetricsServer provides HTTP endpoints for metrics
type MetricsServer struct {
    collector *Collector
}

func NewServer(collector *Collector) *MetricsServer {
    return &MetricsServer{collector: collector}
}

func (s *MetricsServer) ListenAndServe(addr string) error {
    mux := http.NewServeMux()
    
    // Endpoint for JSON metrics
    mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        stats := s.collector.GetStats()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(stats)
    })

    // Endpoint for human-readable metrics
    mux.HandleFunc("/metrics/html", func(w http.ResponseWriter, r *http.Request) {
        stats := s.collector.GetStats()
        w.Header().Set("Content-Type", "text/html")
        s.writeHTMLMetrics(w, stats)
    })

    return http.ListenAndServe(addr, mux)
}

func (s *MetricsServer) writeHTMLMetrics(w http.ResponseWriter, stats *Stats) {
    uptime := time.Since(stats.StartTime)
    fmt.Fprintf(w, "<html><body>")
    fmt.Fprintf(w, "<h1>Reddit Clone Metrics</h1>")
    fmt.Fprintf(w, "<h2>General Statistics</h2>")
    fmt.Fprintf(w, "<ul>")
    fmt.Fprintf(w, "<li>Uptime: %v</li>", uptime)
    fmt.Fprintf(w, "<li>Total Users: %d</li>", stats.TotalUsers)
    fmt.Fprintf(w, "<li>Active Users: %d</li>", stats.ActiveUsers)
    fmt.Fprintf(w, "<li>Request Rate: %.2f/sec</li>", stats.RequestRate)
    fmt.Fprintf(w, "<li>Average Latency: %v</li>", stats.AverageLatency)
    fmt.Fprintf(w, "</ul>")

    fmt.Fprintf(w, "<h2>Content Statistics</h2>")
    fmt.Fprintf(w, "<ul>")
    fmt.Fprintf(w, "<li>Total Posts: %d</li>", stats.TotalPosts)
    fmt.Fprintf(w, "<li>Total Comments: %d</li>", stats.TotalComments)
    fmt.Fprintf(w, "<li>Total Votes: %d</li>", stats.TotalVotes)
    fmt.Fprintf(w, "</ul>")

    fmt.Fprintf(w, "<h2>Endpoint Statistics</h2>")
    fmt.Fprintf(w, "<table border='1'>")
    fmt.Fprintf(w, "<tr><th>Endpoint</th><th>Calls</th><th>Errors</th><th>Avg Latency</th><th>Last Call</th></tr>")
    for _, stat := range stats.EndpointStats {
        fmt.Fprintf(w, "<tr><td>%s</td><td>%d</td><td>%d</td><td>%v</td><td>%v</td></tr>",
            stat.Method, stat.CallCount, stat.ErrorCount, stat.AverageLatency, stat.LastCall.Format(time.RFC3339))
    }
    fmt.Fprintf(w, "</table>")

    fmt.Fprintf(w, "<h2>Subreddit Statistics</h2>")
    fmt.Fprintf(w, "<table border='1'>")
    fmt.Fprintf(w, "<tr><th>Name</th><th>Members</th><th>Posts</th><th>Comments</th><th>Votes</th><th>Active Users</th></tr>")
    for _, stat := range stats.SubredditStats {
        fmt.Fprintf(w, "<tr><td>%s</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td></tr>",
            stat.Name, stat.MemberCount, stat.PostCount, stat.CommentCount, stat.VoteCount, stat.ActiveUsers)
    }
    fmt.Fprintf(w, "</table>")
    fmt.Fprintf(w, "</body></html>")
}