// internal/rest/server.go
package rest

import (
    "encoding/json"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    
    "reddit-clone/api/v1"
    "reddit-clone/internal/engine"
    "reddit-clone/internal/middleware"
)

type Server struct {
    engine *engine.RedditEngine
    router *mux.Router
}

func NewServer(engine *engine.RedditEngine) *Server {
    server := &Server{
        engine: engine,
        router: mux.NewRouter(),
    }
    server.setupRoutes()
    return server
}

func (s *Server) setupRoutes() {
    // Public routes
    s.router.HandleFunc("/api/v1/users/register", s.handleRegister).Methods("POST")
    s.router.HandleFunc("/api/v1/users/login", s.handleLogin).Methods("POST")

    // Protected routes
    // Subreddit routes
    s.router.HandleFunc("/api/v1/subreddits", middleware.AuthMiddleware(s.handleCreateSubreddit)).Methods("POST")
    s.router.HandleFunc("/api/v1/subreddits/{id}", middleware.AuthMiddleware(s.handleGetSubreddit)).Methods("GET")
    s.router.HandleFunc("/api/v1/subreddits", middleware.AuthMiddleware(s.handleListSubreddits)).Methods("GET")
    s.router.HandleFunc("/api/v1/subreddits/{id}/join", middleware.AuthMiddleware(s.handleJoinSubreddit)).Methods("POST")
    s.router.HandleFunc("/api/v1/subreddits/{id}/leave", middleware.AuthMiddleware(s.handleLeaveSubreddit)).Methods("POST")

    // Post routes
    s.router.HandleFunc("/api/v1/posts", middleware.AuthMiddleware(s.handleCreatePost)).Methods("POST")
    s.router.HandleFunc("/api/v1/posts/{id}", middleware.AuthMiddleware(s.handleGetPost)).Methods("GET")
    s.router.HandleFunc("/api/v1/posts", middleware.AuthMiddleware(s.handleListPosts)).Methods("GET")
    s.router.HandleFunc("/api/v1/posts/{id}/vote", middleware.AuthMiddleware(s.handleVote)).Methods("POST")

    // Comment routes
    s.router.HandleFunc("/api/v1/posts/{id}/comments", middleware.AuthMiddleware(s.handleCreateComment)).Methods("POST")
    s.router.HandleFunc("/api/v1/posts/{id}/comments", middleware.AuthMiddleware(s.handleGetComments)).Methods("GET")
    s.router.HandleFunc("/api/v1/comments/{id}/vote", middleware.AuthMiddleware(s.handleVoteComment)).Methods("POST")

    // Feed routes
    s.router.HandleFunc("/api/v1/feed", middleware.AuthMiddleware(s.handleGetFeed)).Methods("GET")

    // Message routes
    s.router.HandleFunc("/api/v1/messages", middleware.AuthMiddleware(s.handleSendMessage)).Methods("POST")
    s.router.HandleFunc("/api/v1/messages", middleware.AuthMiddleware(s.handleGetMessages)).Methods("GET")
    s.router.HandleFunc("/api/v1/messages/{id}", middleware.AuthMiddleware(s.handleGetMessage)).Methods("GET")

    // User routes
    s.router.HandleFunc("/api/v1/users/{id}/public-key", middleware.AuthMiddleware(s.handleGetPublicKey)).Methods("GET") // For bonus feature

    // Add CORS middleware
    s.router.Use(middleware.CORSMiddleware)
}

func (s *Server) Start(port string) error {
    log.Printf("Starting REST server on port %s\n", port)
    return http.ListenAndServe(port, s.router)
}

// Helper methods for responses
func respondWithError(w http.ResponseWriter, code int, message string) {
    respondWithJSON(w, code, api.ErrorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, err := json.Marshal(payload)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("Internal Server Error"))
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}

// Login handler (new)
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    token, err := s.engine.AuthenticateUser(req.Username, req.Password)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
        return
    }

    respondWithJSON(w, http.StatusOK, map[string]string{
        "token": token,
    })
}

// Additional handler for getting a subreddit
func (s *Server) handleGetSubreddit(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    subredditID := vars["id"]

    subreddit, err := s.engine.GetSubReddit(subredditID)
    if err != nil {
        respondWithError(w, http.StatusNotFound, "Subreddit not found")
        return
    }

    resp := api.SubredditResponse{
        ID:          subreddit.ID,
        Name:        subreddit.Name,
        Description: subreddit.Description,
        MemberCount: subreddit.MemberCount,
        CreatedAt:   subreddit.CreatedAt,
    }
    respondWithJSON(w, http.StatusOK, resp)
}

// Handler for listing subreddits
func (s *Server) handleListSubreddits(w http.ResponseWriter, r *http.Request) {
    subreddits, err := s.engine.ListSubreddits()
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to list subreddits")
        return
    }

    var resp []api.SubredditResponse
    for _, sr := range subreddits {
        resp = append(resp, api.SubredditResponse{
            ID:          sr.ID,
            Name:        sr.Name,
            Description: sr.Description,
            MemberCount: sr.MemberCount,
            CreatedAt:   sr.CreatedAt,
        })
    }
    respondWithJSON(w, http.StatusOK, resp)
}

// Handler for listing posts
func (s *Server) handleListPosts(w http.ResponseWriter, r *http.Request) {
    subredditID := r.URL.Query().Get("subreddit_id")
    posts, err := s.engine.ListPosts(subredditID)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to list posts")
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

// Handler for getting comments
func (s *Server) handleGetComments(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    postID := vars["id"]

    comments, err := s.engine.GetComments(postID)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to get comments")
        return
    }

    respondWithJSON(w, http.StatusOK, comments)
}

// Handler for getting public key (bonus feature)
func (s *Server) handleGetPublicKey(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    userID := vars["id"]

    publicKey, err := s.engine.GetUserPublicKey(userID)
    if err != nil {
        respondWithError(w, http.StatusNotFound, "Public key not found")
        return
    }

    respondWithJSON(w, http.StatusOK, map[string]string{
        "public_key": publicKey,
    })
}

// Handler for voting on comments
func (s *Server) handleVoteComment(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    commentID := vars["id"]
    userID := r.Context().Value("userID").(string)

    var req api.VoteRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    err := s.engine.Vote(userID, commentID, req.IsUpvote)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// Handler for getting a single message
func (s *Server) handleGetMessage(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    messageID := vars["id"]
    userID := r.Context().Value("userID").(string)

    message, err := s.engine.GetMessage(userID, messageID)
    if err != nil {
        respondWithError(w, http.StatusNotFound, "Message not found")
        return
    }

    respondWithJSON(w, http.StatusOK, message)
}