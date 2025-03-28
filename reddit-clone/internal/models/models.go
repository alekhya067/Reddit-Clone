// internal/models/models.go
package models

import (
    "time"
    "sync"
)

// User represents a Reddit user
type User struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Password  string    `json:"-"` // Password hash, not exposed in JSON
    Karma     int64     `json:"karma"`
    IsOnline  bool      `json:"is_online"`
    CreatedAt time.Time `json:"created_at"`
}

// SubReddit represents a subreddit
type SubReddit struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    CreatorID   string    `json:"creator_id"`
    MemberCount int64     `json:"member_count"`
    PostCount   int64     `json:"post_count"`
    CreatedAt   time.Time `json:"created_at"`
    Members     sync.Map  `json:"-"` // map[userID]bool
}

// Post represents a post in a subreddit
type Post struct {
    ID           string    `json:"id"`
    Title        string    `json:"title"`
    Content      string    `json:"content"`
    AuthorID     string    `json:"author_id"`
    SubRedditID  string    `json:"subreddit_id"`
    IsRepost     bool      `json:"is_repost"`
    OriginalID   string    `json:"original_id,omitempty"`
    Upvotes      int64     `json:"upvotes"`
    Downvotes    int64     `json:"downvotes"`
    CommentCount int64     `json:"comment_count"`
    CreatedAt    time.Time `json:"created_at"`
}

// Comment represents a comment on a post or another comment
type Comment struct {
    ID        string    `json:"id"`
    Content   string    `json:"content"`
    AuthorID  string    `json:"author_id"`
    PostID    string    `json:"post_id"`
    ParentID  *string   `json:"parent_id"` // nil if top-level comment
    Depth     int       `json:"depth"`     // Comment hierarchy level
    Upvotes   int64     `json:"upvotes"`
    Downvotes int64     `json:"downvotes"`
    CreatedAt time.Time `json:"created_at"`
}

// DirectMessage represents a private message between users
type DirectMessage struct {
    ID        string    `json:"id"`
    FromID    string    `json:"from_id"`
    ToID      string    `json:"to_id"`
    Content   string    `json:"content"`
    IsRead    bool      `json:"is_read"`
    CreatedAt time.Time `json:"created_at"`
}

// Vote represents a user's vote on a post or comment
type Vote struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    TargetID  string    `json:"target_id"` // Post or Comment ID
    IsUpvote  bool      `json:"is_upvote"`
    CreatedAt time.Time `json:"created_at"`
}

// Metrics represents performance and usage metrics
type Metrics struct {
    TotalUsers        int64
    ActiveUsers       int64
    TotalPosts       int64
    TotalComments    int64
    TotalVotes       int64
    AverageLatency   time.Duration
    ResponseTimes    []time.Duration
    StartTime        time.Time
    SubredditStats   map[string]*SubredditMetrics
}

// SubredditMetrics represents metrics for a specific subreddit
type SubredditMetrics struct {
    Name         string
    MemberCount  int64
    PostCount    int64
    CommentCount int64
    VoteCount    int64
    ActiveUsers  int64
}