// internal/client/client.go
package client

import (
    "context"
    "time"
    "sync"
    "errors"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    
    "reddit-clone/internal/models"
    "reddit-clone/internal/proto"
)

type RedditClient struct {
    conn      *grpc.ClientConn
    client    proto.RedditServiceClient
    ctx       context.Context
    cancel    context.CancelFunc
    metrics   *models.Metrics
    mtx       sync.RWMutex
}

func NewRedditClient(serverAddr string) (*RedditClient, error) {
    ctx, cancel := context.WithCancel(context.Background())
    
    // Set up connection with retry
    var conn *grpc.ClientConn
    var err error
    for i := 0; i < 3; i++ {
        conn, err = grpc.DialContext(ctx, serverAddr, 
            grpc.WithInsecure(),
            grpc.WithBlock(),
            grpc.WithTimeout(5*time.Second))
        if err == nil {
            break
        }
        time.Sleep(time.Second)
    }
    if err != nil {
        cancel()
        return nil, err
    }

    return &RedditClient{
        conn:    conn,
        client:  proto.NewRedditServiceClient(conn),
        ctx:     ctx,
        cancel:  cancel,
        metrics: &models.Metrics{
            StartTime:      time.Now(),
            ResponseTimes:  make([]time.Duration, 0),
            SubredditStats: make(map[string]*models.SubredditMetrics),
        },
    }, nil
}

func (c *RedditClient) Close() error {
    c.cancel()
    return c.conn.Close()
}

// RegisterAccount creates a new user account
func (c *RedditClient) RegisterAccount(username, password string) (*models.User, error) {
    start := time.Now()
    resp, err := c.client.RegisterAccount(c.ctx, &proto.RegisterRequest{
        Username: username,
        Password: password,
    })
    
    c.recordLatency(time.Since(start))
    
    if err != nil {
        return nil, handleError(err)
    }

    return &models.User{
        ID:        resp.Id,
        Username:  resp.Username,
        Karma:     resp.Karma,
        IsOnline:  resp.IsOnline,
        CreatedAt: time.Unix(resp.CreatedAt, 0),
    }, nil
}

// CreateSubreddit creates a new subreddit
func (c *RedditClient) CreateSubReddit(name, description, creatorID string) (*models.SubReddit, error) {
    start := time.Now()
    resp, err := c.client.CreateSubreddit(c.ctx, &proto.SubredditRequest{
        Name:        name,
        Description: description,
        CreatorId:   creatorID,
    })
    
    c.recordLatency(time.Since(start))
    
    if err != nil {
        return nil, handleError(err)
    }

    return &models.SubReddit{
        ID:          resp.Id,
        Name:        resp.Name,
        Description: resp.Description,
        CreatorID:   resp.CreatorId,
        MemberCount: resp.MemberCount,
        CreatedAt:   time.Unix(resp.CreatedAt, 0),
        Members:     sync.Map{},
    }, nil
}

// Continue with all other methods...
// Continuing internal/client/client.go...

// JoinSubReddit adds a user to a subreddit
func (c *RedditClient) JoinSubReddit(userID, subredditID string) error {
    start := time.Now()
    _, err := c.client.JoinSubreddit(c.ctx, &proto.JoinRequest{
        UserId:      userID,
        SubredditId: subredditID,
    })
    
    c.recordLatency(time.Since(start))
    return handleError(err)
}

// LeaveSubReddit removes a user from a subreddit
func (c *RedditClient) LeaveSubReddit(userID, subredditID string) error {
    start := time.Now()
    _, err := c.client.LeaveSubreddit(c.ctx, &proto.JoinRequest{
        UserId:      userID,
        SubredditId: subredditID,
    })
    
    c.recordLatency(time.Since(start))
    return handleError(err)
}

// CreatePost creates a new post in a subreddit
func (c *RedditClient) CreatePost(title, content, authorID, subredditID string) (*models.Post, error) {
    start := time.Now()
    resp, err := c.client.CreatePost(c.ctx, &proto.PostRequest{
        Title:       title,
        Content:     content,
        AuthorId:    authorID,
        SubredditId: subredditID,
    })
    
    c.recordLatency(time.Since(start))
    
    if err != nil {
        return nil, handleError(err)
    }

    return &models.Post{
        ID:          resp.Id,
        Title:       resp.Title,
        Content:     resp.Content,
        AuthorID:    resp.AuthorId,
        SubRedditID: resp.SubredditId,
        Upvotes:     resp.Upvotes,
        Downvotes:   resp.Downvotes,
        CreatedAt:   time.Unix(resp.CreatedAt, 0),
    }, nil
}

// CreateComment adds a comment to a post or another comment
func (c *RedditClient) CreateComment(content, authorID, postID string, parentCommentID *string) (*models.Comment, error) {
    start := time.Now()
    req := &proto.CommentRequest{
        Content:   content,
        AuthorId:  authorID,
        PostId:    postID,
        ParentId:  parentCommentID,  // This is already a *string
    }
    
    resp, err := c.client.CreateComment(c.ctx, req)
    c.recordLatency(time.Since(start))
    
    if err != nil {
        return nil, handleError(err)
    }

    // Simplified handling of optional ParentID
    var parentID *string
    if resp.ParentId != "" {  // Change this line
        temp := resp.ParentId  // And this line
        parentID = &temp
    }

    return &models.Comment{
        ID:        resp.Id,
        Content:   resp.Content,
        AuthorID:  resp.AuthorId,
        PostID:    resp.PostId,
        ParentID:  parentID,
        Depth:     int(resp.Depth),
        Upvotes:   resp.Upvotes,
        Downvotes: resp.Downvotes,
        CreatedAt: time.Unix(resp.CreatedAt, 0),
    }, nil
}

// Vote handles upvoting and downvoting of posts and comments
func (c *RedditClient) Vote(userID, targetID string, isUpvote bool) error {
    start := time.Now()
    _, err := c.client.Vote(c.ctx, &proto.VoteRequest{
        UserId:    userID,
        TargetId:  targetID,
        IsUpvote:  isUpvote,
    })
    
    c.recordLatency(time.Since(start))
    return handleError(err)
}

// GetFeed returns a list of posts from subscribed subreddits
func (c *RedditClient) GetFeed(userID string) ([]*models.Post, error) {
    start := time.Now()
    resp, err := c.client.GetFeed(c.ctx, &proto.FeedRequest{
        UserId: userID,
    })
    
    c.recordLatency(time.Since(start))
    
    if err != nil {
        return nil, handleError(err)
    }

    posts := make([]*models.Post, len(resp.Posts))
    for i, p := range resp.Posts {
        posts[i] = &models.Post{
            ID:          p.Id,
            Title:       p.Title,
            Content:     p.Content,
            AuthorID:    p.AuthorId,
            SubRedditID: p.SubredditId,
            Upvotes:     p.Upvotes,
            Downvotes:   p.Downvotes,
            CreatedAt:   time.Unix(p.CreatedAt, 0),
        }
    }
    return posts, nil
}

// SendDirectMessage sends a message from one user to another
func (c *RedditClient) SendDirectMessage(fromID, toID, content string) (*models.DirectMessage, error) {
    start := time.Now()
    resp, err := c.client.SendMessage(c.ctx, &proto.MessageRequest{
        FromId:  fromID,
        ToId:    toID,
        Content: content,
    })
    
    c.recordLatency(time.Since(start))
    
    if err != nil {
        return nil, handleError(err)
    }

    return &models.DirectMessage{
        ID:        resp.Id,
        FromID:    resp.FromId,
        ToID:      resp.ToId,
        Content:   resp.Content,
        IsRead:    resp.IsRead,
        CreatedAt: time.Unix(resp.CreatedAt, 0),
    }, nil
}

// GetUserMessages returns all messages for a user
func (c *RedditClient) GetUserMessages(userID string) ([]*models.DirectMessage, error) {
    start := time.Now()
    resp, err := c.client.GetUserMessages(c.ctx, &proto.UserRequest{
        UserId: userID,
    })
    
    c.recordLatency(time.Since(start))
    
    if err != nil {
        return nil, handleError(err)
    }

    messages := make([]*models.DirectMessage, len(resp.Messages))
    for i, m := range resp.Messages {
        messages[i] = &models.DirectMessage{
            ID:        m.Id,
            FromID:    m.FromId,
            ToID:      m.ToId,
            Content:   m.Content,
            IsRead:    m.IsRead,
            CreatedAt: time.Unix(m.CreatedAt, 0),
        }
    }
    return messages, nil
}

// Helper methods for metrics and error handling
func (c *RedditClient) recordLatency(duration time.Duration) {
    c.mtx.Lock()
    defer c.mtx.Unlock()
    c.metrics.ResponseTimes = append(c.metrics.ResponseTimes, duration)
}

func (c *RedditClient) GetMetrics() *models.Metrics {
    c.mtx.RLock()
    defer c.mtx.RUnlock()
    
    // Calculate average latency
    var totalLatency time.Duration
    for _, latency := range c.metrics.ResponseTimes {
        totalLatency += latency
    }
    if len(c.metrics.ResponseTimes) > 0 {
        c.metrics.AverageLatency = totalLatency / time.Duration(len(c.metrics.ResponseTimes))
    }
    
    return c.metrics
}

// Error handling helper
func handleError(err error) error {
    if err == nil {
        return nil
    }

    st, ok := status.FromError(err)
    if !ok {
        return err
    }

    switch st.Code() {
    case codes.NotFound:
        return errors.New(st.Message())
    case codes.AlreadyExists:
        return errors.New(st.Message())
    case codes.PermissionDenied:
        return errors.New(st.Message())
    case codes.Unavailable:
        return errors.New("service temporarily unavailable")
    default:
        return err
    }
}