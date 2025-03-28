// api/v1/api.go
package api

import "time"

// Request types
type RegisterRequest struct {
    Username  string `json:"username"`
    Password  string `json:"password"`
    PublicKey string `json:"public_key,omitempty"` // For bonus feature
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Token string `json:"token"`
}

type SubredditRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type PostRequest struct {
    Title       string `json:"title"`
    Content     string `json:"content"`
    SubredditID string `json:"subreddit_id"`
    Signature   string `json:"signature,omitempty"` // For bonus feature
}

type CommentRequest struct {
    Content    string  `json:"content"`
    PostID     string  `json:"post_id"`
    ParentID   *string `json:"parent_id,omitempty"`
}

type VoteRequest struct {
    IsUpvote bool `json:"is_upvote"`
}

type MessageRequest struct {
    ToID    string `json:"to_id"`
    Content string `json:"content"`
}

// Response types
type UserResponse struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Karma     int64     `json:"karma"`
    CreatedAt time.Time `json:"created_at"`
}

type SubredditResponse struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    MemberCount int64     `json:"member_count"`
    CreatorID   string    `json:"creator_id"`
    CreatedAt   time.Time `json:"created_at"`
}

type PostResponse struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Content     string    `json:"content"`
    AuthorID    string    `json:"author_id"`
    SubredditID string    `json:"subreddit_id"`
    Upvotes     int64     `json:"upvotes"`
    Downvotes   int64     `json:"downvotes"`
    CommentCount int64    `json:"comment_count"`
    CreatedAt   time.Time `json:"created_at"`
    Signature   string    `json:"signature,omitempty"` // For bonus feature
}

type CommentResponse struct {
    ID        string    `json:"id"`
    Content   string    `json:"content"`
    AuthorID  string    `json:"author_id"`
    PostID    string    `json:"post_id"`
    ParentID  *string   `json:"parent_id"`
    Depth     int32     `json:"depth"`
    Upvotes   int64     `json:"upvotes"`
    Downvotes int64     `json:"downvotes"`
    CreatedAt time.Time `json:"created_at"`
}

type MessageResponse struct {
    ID        string    `json:"id"`
    FromID    string    `json:"from_id"`
    ToID      string    `json:"to_id"`
    Content   string    `json:"content"`
    IsRead    bool      `json:"is_read"`
    CreatedAt time.Time `json:"created_at"`
}

type FeedResponse struct {
    Posts []PostResponse `json:"posts"`
}

type StatusResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
    Error   string `json:"error"`
    Code    int    `json:"code,omitempty"`
    Details string `json:"details,omitempty"`
}

// List response types
type SubredditListResponse struct {
    Subreddits []SubredditResponse `json:"subreddits"`
    Total      int                 `json:"total"`
}

type PostListResponse struct {
    Posts []PostResponse `json:"posts"`
    Total int           `json:"total"`
}

type CommentListResponse struct {
    Comments []CommentResponse `json:"comments"`
    Total    int              `json:"total"`
}

type MessageListResponse struct {
    Messages []MessageResponse `json:"messages"`
    Total    int              `json:"total"`
}

// Search request/response
type SearchRequest struct {
    Query       string `json:"query"`
    SubredditID string `json:"subreddit_id,omitempty"`
    Type        string `json:"type,omitempty"` // "posts", "comments", "subreddits"
    Page        int    `json:"page,omitempty"`
    Limit       int    `json:"limit,omitempty"`
}

type SearchResponse struct {
    Posts      []PostResponse      `json:"posts,omitempty"`
    Comments   []CommentResponse   `json:"comments,omitempty"`
    Subreddits []SubredditResponse `json:"subreddits,omitempty"`
    Total      int                 `json:"total"`
}

// Pagination params
type PaginationParams struct {
    Page  int `json:"page"`
    Limit int `json:"limit"`
}

// Sort params
type SortParams struct {
    Field     string `json:"field"`
    Direction string `json:"direction"` // "asc" or "desc"
}