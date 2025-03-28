// internal/web/client.go
package web

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "reddit-clone/api/v1"
)

type Client struct {
    baseURL    string
    httpClient *http.Client
    token      string
}

func NewClient(baseURL string) *Client {
    return &Client{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: time.Second * 10,
        },
    }
}

func (c *Client) SetToken(token string) {
    c.token = token
}

// Authentication methods
func (c *Client) Register(username, password string) error {
    req := api.RegisterRequest{
        Username: username,
        Password: password,
    }
    return c.post("/api/v1/users/register", req, nil)
}

func (c *Client) Login(username, password string) (string, error) {
    req := struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }{
        Username: username,
        Password: password,
    }
    
    var resp struct {
        Token string `json:"token"`
    }
    
    err := c.post("/api/v1/users/login", req, &resp)
    if err != nil {
        return "", err
    }
    
    c.token = resp.Token
    return resp.Token, nil
}

// Subreddit methods
func (c *Client) CreateSubreddit(name, description string) (*api.SubredditResponse, error) {
    req := api.SubredditRequest{
        Name:        name,
        Description: description,
    }
    
    var resp api.SubredditResponse
    err := c.post("/api/v1/subreddits", req, &resp)
    if err != nil {
        return nil, err
    }
    return &resp, nil
}

func (c *Client) JoinSubreddit(subredditID string) error {
    return c.post(fmt.Sprintf("/api/v1/subreddits/%s/join", subredditID), nil, nil)
}

func (c *Client) LeaveSubreddit(subredditID string) error {
    return c.post(fmt.Sprintf("/api/v1/subreddits/%s/leave", subredditID), nil, nil)
}

// Post methods
func (c *Client) CreatePost(title, content, subredditID string) (*api.PostResponse, error) {
    req := api.PostRequest{
        Title:       title,
        Content:     content,
        SubredditID: subredditID,
    }
    
    var resp api.PostResponse
    err := c.post("/api/v1/posts", req, &resp)
    if err != nil {
        return nil, err
    }
    return &resp, nil
}

func (c *Client) GetPost(postID string) (*api.PostResponse, error) {
    var resp api.PostResponse
    err := c.get(fmt.Sprintf("/api/v1/posts/%s", postID), &resp)
    if err != nil {
        return nil, err
    }
    return &resp, nil
}

func (c *Client) GetFeed() ([]api.PostResponse, error) {
    var resp []api.PostResponse
    err := c.get("/api/v1/feed", &resp)
    if err != nil {
        return nil, err
    }
    if resp == nil {
        resp = make([]api.PostResponse, 0)
    }
    return resp, nil
}

// Comment methods
func (c *Client) CreateComment(content, postID string, parentID *string) (*api.CommentResponse, error) {
    req := api.CommentRequest{
        Content:  content,
        PostID:   postID,
        ParentID: parentID,
    }
    
    var resp api.CommentResponse
    err := c.post(fmt.Sprintf("/api/v1/posts/%s/comments", postID), req, &resp)
    if err != nil {
        return nil, err
    }
    return &resp, nil
}

// Vote methods
func (c *Client) Vote(targetID string, isUpvote bool) error {
    req := api.VoteRequest{
        IsUpvote: isUpvote,
    }
    return c.post(fmt.Sprintf("/api/v1/posts/%s/vote", targetID), req, nil)
}

// Message methods
func (c *Client) SendMessage(toID, content string) (*api.MessageResponse, error) {
    req := struct {
        ToID    string `json:"to_id"`
        Content string `json:"content"`
    }{
        ToID:    toID,
        Content: content,
    }
    
    var resp api.MessageResponse
    err := c.post("/api/v1/messages", req, &resp)
    if err != nil {
        return nil, err
    }
    return &resp, nil
}

func (c *Client) GetMessages() ([]api.MessageResponse, error) {
    var resp []api.MessageResponse
    err := c.get("/api/v1/messages", &resp)
    if err != nil {
        return nil, err
    }
    if resp == nil {
        resp = make([]api.MessageResponse, 0)
    }
    return resp, nil
}

// Helper methods
func (c *Client) get(path string, response interface{}) error {
    return c.doRequest(http.MethodGet, path, nil, response)
}

func (c *Client) post(path string, body interface{}, response interface{}) error {
    return c.doRequest(http.MethodPost, path, body, response)
}

func (c *Client) doRequest(method, path string, body interface{}, response interface{}) error {
    var bodyReader *bytes.Reader
    
    if body != nil {
        bodyBytes, err := json.Marshal(body)
        if err != nil {
            return fmt.Errorf("failed to marshal request body: %w", err)
        }
        bodyReader = bytes.NewReader(bodyBytes)
    }

    url := fmt.Sprintf("%s%s", c.baseURL, path)
    var req *http.Request
    var err error
    
    if bodyReader != nil {
        req, err = http.NewRequest(method, url, bodyReader)
    } else {
        req, err = http.NewRequest(method, url, nil)
    }
    
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    if c.token != "" {
        req.Header.Set("Authorization", "Bearer "+c.token)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        var errResp api.ErrorResponse
        if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
            return fmt.Errorf("request failed with status %d", resp.StatusCode)
        }
        return fmt.Errorf("request failed: %s", errResp.Error)
    }

    if response != nil {
        if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
            return fmt.Errorf("failed to decode response: %w", err)
        }
    }

    return nil
}