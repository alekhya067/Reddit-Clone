// internal/rest/handlers.go
package rest

import (
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
    
    "reddit-clone/api/v1"
)

// User handlers
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
    var req api.RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    user, err := s.engine.RegisterAccount(req.Username, req.Password)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, err.Error())
        return
    }

    resp := api.UserResponse{
        ID:        user.ID,
        Username:  user.Username,
        Karma:     user.Karma,
        CreatedAt: user.CreatedAt,
    }
    respondWithJSON(w, http.StatusCreated, resp)
}

// Subreddit handlers
func (s *Server) handleCreateSubreddit(w http.ResponseWriter, r *http.Request) {
    var req api.SubredditRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    // Get user ID from context (after implementing auth middleware)
    userID := r.Context().Value("userID").(string)

    subreddit, err := s.engine.CreateSubReddit(req.Name, req.Description, userID)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, err.Error())
        return
    }

    resp := api.SubredditResponse{
        ID:          subreddit.ID,
        Name:        subreddit.Name,
        Description: subreddit.Description,
        MemberCount: subreddit.MemberCount,
        CreatedAt:   subreddit.CreatedAt,
    }
    respondWithJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleJoinSubreddit(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    subredditID := vars["id"]
    userID := r.Context().Value("userID").(string)

    err := s.engine.JoinSubReddit(userID, subredditID)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (s *Server) handleLeaveSubreddit(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    subredditID := vars["id"]
    userID := r.Context().Value("userID").(string)

    err := s.engine.LeaveSubReddit(userID, subredditID)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// Post handlers
func (s *Server) handleCreatePost(w http.ResponseWriter, r *http.Request) {
    var req api.PostRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    userID := r.Context().Value("userID").(string)

    post, err := s.engine.CreatePost(req.Title, req.Content, userID, req.SubredditID)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, err.Error())
        return
    }

    resp := api.PostResponse{
        ID:          post.ID,
        Title:       post.Title,
        Content:     post.Content,
        AuthorID:    post.AuthorID,
        SubredditID: post.SubRedditID,
        Upvotes:     post.Upvotes,
        Downvotes:   post.Downvotes,
        CreatedAt:   post.CreatedAt,
    }
    respondWithJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleGetPost(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    postID := vars["id"]

    // Add method to engine to get single post
    post, err := s.engine.GetPost(postID)
    if err != nil {
        respondWithError(w, http.StatusNotFound, "Post not found")
        return
    }

    resp := api.PostResponse{
        ID:          post.ID,
        Title:       post.Title,
        Content:     post.Content,
        AuthorID:    post.AuthorID,
        SubredditID: post.SubRedditID,
        Upvotes:     post.Upvotes,
        Downvotes:   post.Downvotes,
        CreatedAt:   post.CreatedAt,
    }
    respondWithJSON(w, http.StatusOK, resp)
}

func (s *Server) handleVote(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    targetID := vars["id"]
    userID := r.Context().Value("userID").(string)

    var req api.VoteRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    err := s.engine.Vote(userID, targetID, req.IsUpvote)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// Feed handler
func (s *Server) handleGetFeed(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)

    posts, err := s.engine.GetFeed(userID)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    var resp []api.PostResponse
    for _, post := range posts {
        resp = append(resp, api.PostResponse{
            ID:          post.ID,
            Title:       post.Title,
            Content:     post.Content,
            AuthorID:    post.AuthorID,
            SubredditID: post.SubRedditID,
            Upvotes:     post.Upvotes,
            Downvotes:   post.Downvotes,
            CreatedAt:   post.CreatedAt,
        })
    }
    respondWithJSON(w, http.StatusOK, resp)
}

// Message handlers
func (s *Server) handleGetMessages(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)

    messages, err := s.engine.GetUserMessages(userID)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, messages)
}

func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)

    var req struct {
        ToID    string `json:"to_id"`
        Content string `json:"content"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    message, err := s.engine.SendDirectMessage(userID, req.ToID, req.Content)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, err.Error())
        return
    }

    respondWithJSON(w, http.StatusCreated, message)
}

// Add this to internal/rest/handlers.go
func (s *Server) handleCreateComment(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    postID := vars["id"]
    userID := r.Context().Value("userID").(string)

    var req api.CommentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    comment, err := s.engine.CreateComment(
        req.Content,
        userID,
        postID,
        req.ParentID,
    )
    if err != nil {
        respondWithError(w, http.StatusBadRequest, err.Error())
        return
    }

    resp := api.CommentResponse{
        ID:        comment.ID,
        Content:   comment.Content,
        AuthorID:  comment.AuthorID,
        PostID:    comment.PostID,
        ParentID:  comment.ParentID,
        Upvotes:   comment.Upvotes,
        Downvotes: comment.Downvotes,
        CreatedAt: comment.CreatedAt,
    }
    respondWithJSON(w, http.StatusCreated, resp)
}