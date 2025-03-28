// internal/server/server.go
package server

import (
    "context"
    "time"
    "reddit-clone/internal/engine"
    "reddit-clone/internal/proto"
    "reddit-clone/pkg/metrics"
)

type RedditServer struct {
    proto.UnimplementedRedditServiceServer
    engine  *engine.RedditEngine
    metrics *metrics.Collector
}

func NewRedditServer(engine *engine.RedditEngine, metrics *metrics.Collector) *RedditServer {
    return &RedditServer{
        engine:  engine,
        metrics: metrics,
    }
}

// RegisterAccount handles user registration
func (s *RedditServer) RegisterAccount(ctx context.Context, req *proto.RegisterRequest) (*proto.UserResponse, error) {
    start := time.Now()
    defer func() {
        s.metrics.RecordLatency("RegisterAccount", time.Since(start))
    }()

    user, err := s.engine.RegisterAccount(req.Username, req.Password)
    if err != nil {
        s.metrics.RecordError("RegisterAccount")
        return nil, err
    }

    return &proto.UserResponse{
        Id:        user.ID,
        Username:  user.Username,
        Karma:     user.Karma,
        IsOnline:  user.IsOnline,
        CreatedAt: user.CreatedAt.Unix(),
    }, nil
}

// CreateSubreddit handles subreddit creation
func (s *RedditServer) CreateSubreddit(ctx context.Context, req *proto.SubredditRequest) (*proto.SubredditResponse, error) {
    start := time.Now()
    defer func() {
        s.metrics.RecordLatency("CreateSubreddit", time.Since(start))
    }()

    subreddit, err := s.engine.CreateSubReddit(req.Name, req.Description, req.CreatorId)
    if err != nil {
        s.metrics.RecordError("CreateSubreddit")
        return nil, err
    }

    return &proto.SubredditResponse{
        Id:          subreddit.ID,
        Name:        subreddit.Name,
        Description: subreddit.Description,
        CreatorId:   subreddit.CreatorID,
        MemberCount: subreddit.MemberCount,
        CreatedAt:   subreddit.CreatedAt.Unix(),
    }, nil
}

// JoinSubreddit handles joining a subreddit
func (s *RedditServer) JoinSubreddit(ctx context.Context, req *proto.JoinRequest) (*proto.StatusResponse, error) {
    start := time.Now()
    defer func() {
        s.metrics.RecordLatency("JoinSubreddit", time.Since(start))
    }()

    err := s.engine.JoinSubReddit(req.UserId, req.SubredditId)
    if err != nil {
        s.metrics.RecordError("JoinSubreddit")
        return &proto.StatusResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    return &proto.StatusResponse{Success: true}, nil
}

// LeaveSubreddit handles leaving a subreddit
func (s *RedditServer) LeaveSubreddit(ctx context.Context, req *proto.JoinRequest) (*proto.StatusResponse, error) {
    start := time.Now()
    defer func() {
        s.metrics.RecordLatency("LeaveSubreddit", time.Since(start))
    }()

    err := s.engine.LeaveSubReddit(req.UserId, req.SubredditId)
    if err != nil {
        s.metrics.RecordError("LeaveSubreddit")
        return &proto.StatusResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    return &proto.StatusResponse{Success: true}, nil
}

// CreatePost handles post creation
func (s *RedditServer) CreatePost(ctx context.Context, req *proto.PostRequest) (*proto.PostResponse, error) {
    start := time.Now()
    defer func() {
        s.metrics.RecordLatency("CreatePost", time.Since(start))
    }()

    post, err := s.engine.CreatePost(req.Title, req.Content, req.AuthorId, req.SubredditId)
    if err != nil {
        s.metrics.RecordError("CreatePost")
        return nil, err
    }

    return &proto.PostResponse{
        Id:          post.ID,
        Title:       post.Title,
        Content:     post.Content,
        AuthorId:    post.AuthorID,
        SubredditId: post.SubRedditID,
        Upvotes:     post.Upvotes,
        Downvotes:   post.Downvotes,
        CreatedAt:   post.CreatedAt.Unix(),
    }, nil
}

// CreateComment handles comment creation
unc (s *RedditServer) CreateComment(ctx context.Context, req *proto.CommentRequest) (*proto.CommentResponse, error) {
    start := time.Now()
    defer func() {
        s.metrics.RecordLatency("CreateComment", time.Since(start))
    }()

    comment, err := s.engine.CreateComment(req.Content, req.AuthorId, req.PostId, req.ParentId)
    if err != nil {
        s.metrics.RecordError("CreateComment")
        return nil, err
    }

    // Handle the optional ParentID
    var parentId string
    if comment.ParentID != nil {
        parentId = *comment.ParentID
    }

    return &proto.CommentResponse{
        Id:        comment.ID,
        Content:   comment.Content,
        AuthorId:  comment.AuthorID,
        PostId:    comment.PostID,
        ParentId:  parentId,          // Now using string instead of *string
        Depth:     int32(comment.Depth),
        Upvotes:   comment.Upvotes,
        Downvotes: comment.Downvotes,
        CreatedAt: comment.CreatedAt.Unix(),
    }, nil
}

// Vote handles voting on posts and comments
func (s *RedditServer) Vote(ctx context.Context, req *proto.VoteRequest) (*proto.StatusResponse, error) {
    start := time.Now()
    defer func() {
        s.metrics.RecordLatency("Vote", time.Since(start))
    }()

    err := s.engine.Vote(req.UserId, req.TargetId, req.IsUpvote)
    if err != nil {
        s.metrics.RecordError("Vote")
        return &proto.StatusResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    return &proto.StatusResponse{Success: true}, nil
}

// GetFeed handles retrieving a user's feed
func (s *RedditServer) GetFeed(ctx context.Context, req *proto.FeedRequest) (*proto.FeedResponse, error) {
    start := time.Now()
    defer func() {
        s.metrics.RecordLatency("GetFeed", time.Since(start))
    }()

    posts, err := s.engine.GetFeed(req.UserId)
    if err != nil {
        s.metrics.RecordError("GetFeed")
        return nil, err
    }

    protoPosts := make([]*proto.PostResponse, len(posts))
    for i, post := range posts {
        protoPosts[i] = &proto.PostResponse{
            Id:          post.ID,
            Title:       post.Title,
            Content:     post.Content,
            AuthorId:    post.AuthorID,
            SubredditId: post.SubRedditID,
            Upvotes:     post.Upvotes,
            Downvotes:   post.Downvotes,
            CreatedAt:   post.CreatedAt.Unix(),
        }
    }

    return &proto.FeedResponse{Posts: protoPosts}, nil
}

// SendMessage handles sending direct messages
func (s *RedditServer) SendMessage(ctx context.Context, req *proto.MessageRequest) (*proto.MessageResponse, error) {
    start := time.Now()
    defer func() {
        s.metrics.RecordLatency("SendMessage", time.Since(start))
    }()

    msg, err := s.engine.SendDirectMessage(req.FromId, req.ToId, req.Content)
    if err != nil {
        s.metrics.RecordError("SendMessage")
        return nil, err
    }

    return &proto.MessageResponse{
        Id:        msg.ID,
        FromId:    msg.FromID,
        ToId:      msg.ToID,
        Content:   msg.Content,
        IsRead:    msg.IsRead,
        CreatedAt: msg.CreatedAt.Unix(),
    }, nil
}

// GetUserMessages handles retrieving a user's messages
func (s *RedditServer) GetUserMessages(ctx context.Context, req *proto.UserRequest) (*proto.MessagesResponse, error) {
    start := time.Now()
    defer func() {
        s.metrics.RecordLatency("GetUserMessages", time.Since(start))
    }()

    messages, err := s.engine.GetUserMessages(req.UserId)
    if err != nil {
        s.metrics.RecordError("GetUserMessages")
        return nil, err
    }

    protoMessages := make([]*proto.MessageResponse, len(messages))
    for i, msg := range messages {
        protoMessages[i] = &proto.MessageResponse{
            Id:        msg.ID,
            FromId:    msg.FromID,
            ToId:      msg.ToID,
            Content:   msg.Content,
            IsRead:    msg.IsRead,
            CreatedAt: msg.CreatedAt.Unix(),
        }
    }

    return &proto.MessagesResponse{Messages: protoMessages}, nil
}