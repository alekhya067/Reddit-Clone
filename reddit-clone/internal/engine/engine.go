// internal/engine/engine.go
package engine

import (
    "errors"
    "sync"
    "time"
    "crypto/rand"
    "encoding/hex"
    "golang.org/x/crypto/bcrypt"
    
    "reddit-clone/internal/models"
)

type RedditEngine struct {
    users      sync.Map // map[string]*models.User
    subreddits sync.Map // map[string]*models.SubReddit
    posts      sync.Map // map[string]*models.Post
    comments   sync.Map // map[string]*models.Comment
    messages   sync.Map // map[string]*models.DirectMessage
    votes      sync.Map // map[string]*models.Vote
}

func NewRedditEngine() *RedditEngine {
    return &RedditEngine{}
}

func generateID() string {
    bytes := make([]byte, 16)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}

// Start the engine server
func (e *RedditEngine) Start(port string) error {
    return nil
}

// RegisterAccount creates a new user account
func (e *RedditEngine) RegisterAccount(username, password string) (*models.User, error) {
    // Check if username already exists
    var exists bool
    e.users.Range(func(key, value interface{}) bool {
        user := value.(*models.User)
        if user.Username == username {
            exists = true
            return false
        }
        return true
    })

    if exists {
        return nil, errors.New("username already exists")
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    user := &models.User{
        ID:        generateID(),
        Username:  username,
        Password:  string(hashedPassword),
        Karma:     0,
        CreatedAt: time.Now(),
    }

    e.users.Store(user.ID, user)
    return user, nil
}

// AuthenticateUser validates credentials and returns a token
func (e *RedditEngine) AuthenticateUser(username, password string) (string, error) {
    var user *models.User
    e.users.Range(func(key, value interface{}) bool {
        u := value.(*models.User)
        if u.Username == username {
            user = u
            return false
        }
        return true
    })

    if user == nil {
        return "", errors.New("user not found")
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
        return "", errors.New("invalid password")
    }

    return user.ID, nil // Using user ID as token for simplicity
}

// CreateSubReddit creates a new subreddit
func (e *RedditEngine) CreateSubReddit(name, description, creatorID string) (*models.SubReddit, error) {
    // Validate creator exists
    _, exists := e.users.Load(creatorID)
    if !exists {
        return nil, errors.New("creator not found")
    }

    subreddit := &models.SubReddit{
        ID:          generateID(),
        Name:        name,
        Description: description,
        CreatorID:   creatorID,
        CreatedAt:   time.Now(),
        Members:     sync.Map{},
    }

    // Add creator as first member
    subreddit.Members.Store(creatorID, true)
    e.subreddits.Store(subreddit.ID, subreddit)
    return subreddit, nil
}

// GetSubReddit retrieves a subreddit by ID
func (e *RedditEngine) GetSubReddit(subredditID string) (*models.SubReddit, error) {
    subI, ok := e.subreddits.Load(subredditID)
    if !ok {
        return nil, errors.New("subreddit not found")
    }
    return subI.(*models.SubReddit), nil
}

// ListSubreddits returns all subreddits
func (e *RedditEngine) ListSubreddits() ([]*models.SubReddit, error) {
    var subreddits []*models.SubReddit
    e.subreddits.Range(func(key, value interface{}) bool {
        subreddits = append(subreddits, value.(*models.SubReddit))
        return true
    })
    return subreddits, nil
}

// JoinSubReddit adds a user to a subreddit
func (e *RedditEngine) JoinSubReddit(userID, subredditID string) error {
    subredditI, exists := e.subreddits.Load(subredditID)
    if !exists {
        return errors.New("subreddit not found")
    }

    _, exists = e.users.Load(userID)
    if !exists {
        return errors.New("user not found")
    }

    subreddit := subredditI.(*models.SubReddit)
    subreddit.Members.Store(userID, true)
    return nil
}

// LeaveSubReddit removes a user from a subreddit
func (e *RedditEngine) LeaveSubReddit(userID, subredditID string) error {
    subredditI, exists := e.subreddits.Load(subredditID)
    if !exists {
        return errors.New("subreddit not found")
    }

    subreddit := subredditI.(*models.SubReddit)
    subreddit.Members.Delete(userID)
    return nil
}

// CreatePost creates a new post in a subreddit
func (e *RedditEngine) CreatePost(title, content, authorID, subredditID string) (*models.Post, error) {
    // Validate author and subreddit exist
    _, authorExists := e.users.Load(authorID)
    subredditI, subredditExists := e.subreddits.Load(subredditID)

    if !authorExists {
        return nil, errors.New("author not found")
    }
    if !subredditExists {
        return nil, errors.New("subreddit not found")
    }

    // Check if user is a member of the subreddit
    subreddit := subredditI.(*models.SubReddit)
    _, isMember := subreddit.Members.Load(authorID)
    if !isMember {
        return nil, errors.New("user is not a member of this subreddit")
    }

    post := &models.Post{
        ID:          generateID(),
        Title:       title,
        Content:     content,
        AuthorID:    authorID,
        SubRedditID: subredditID,
        CreatedAt:   time.Now(),
    }

    e.posts.Store(post.ID, post)
    return post, nil
}

// GetPost retrieves a single post by ID
func (e *RedditEngine) GetPost(postID string) (*models.Post, error) {
    postI, ok := e.posts.Load(postID)
    if !ok {
        return nil, errors.New("post not found")
    }
    return postI.(*models.Post), nil
}

// ListPosts returns posts for a subreddit
func (e *RedditEngine) ListPosts(subredditID string) ([]*models.Post, error) {
    var posts []*models.Post
    e.posts.Range(func(key, value interface{}) bool {
        post := value.(*models.Post)
        if post.SubRedditID == subredditID {
            posts = append(posts, post)
        }
        return true
    })
    return posts, nil
}

// CreateComment adds a comment to a post or another comment
func (e *RedditEngine) CreateComment(content, authorID, postID string, parentCommentID *string) (*models.Comment, error) {
    // Validate author and post exist
    _, authorExists := e.users.Load(authorID)
    _, postExists := e.posts.Load(postID)

    if !authorExists {
        return nil, errors.New("author not found")
    }
    if !postExists {
        return nil, errors.New("post not found")
    }

    // If parent comment ID is provided, validate it exists
    if parentCommentID != nil {
        _, exists := e.comments.Load(*parentCommentID)
        if !exists {
            return nil, errors.New("parent comment not found")
        }
    }

    comment := &models.Comment{
        ID:        generateID(),
        Content:   content,
        AuthorID:  authorID,
        PostID:    postID,
        ParentID:  parentCommentID,
        CreatedAt: time.Now(),
    }

    e.comments.Store(comment.ID, comment)
    return comment, nil
}

// GetComments returns comments for a post
func (e *RedditEngine) GetComments(postID string) ([]*models.Comment, error) {
    var comments []*models.Comment
    e.comments.Range(func(key, value interface{}) bool {
        comment := value.(*models.Comment)
        if comment.PostID == postID {
            comments = append(comments, comment)
        }
        return true
    })
    return comments, nil
}

// Vote handles upvoting and downvoting of posts and comments
func (e *RedditEngine) Vote(userID, targetID string, isUpvote bool) error {
    // Check if target exists (could be post or comment)
    postI, isPost := e.posts.Load(targetID)
    commentI, isComment := e.comments.Load(targetID)

    if !isPost && !isComment {
        return errors.New("target not found")
    }

    voteID := userID + ":" + targetID
    existingVoteI, exists := e.votes.Load(voteID)

    if exists {
        // Update existing vote
        existingVote := existingVoteI.(*models.Vote)
        if existingVote.IsUpvote != isUpvote {
            if isPost {
                post := postI.(*models.Post)
                if isUpvote {
                    post.Upvotes++
                    post.Downvotes--
                } else {
                    post.Downvotes++
                    post.Upvotes--
                }
            } else {
                comment := commentI.(*models.Comment)
                if isUpvote {
                    comment.Upvotes++
                    comment.Downvotes--
                } else {
                    comment.Downvotes++
                    comment.Upvotes--
                }
            }
            existingVote.IsUpvote = isUpvote
        }
    } else {
        // Create new vote
        vote := &models.Vote{
            UserID:    userID,
            TargetID:  targetID,
            IsUpvote:  isUpvote,
            CreatedAt: time.Now(),
        }

        if isPost {
            post := postI.(*models.Post)
            if isUpvote {
                post.Upvotes++
            } else {
                post.Downvotes++
            }
        } else {
            comment := commentI.(*models.Comment)
            if isUpvote {
                comment.Upvotes++
            } else {
                comment.Downvotes++
            }
        }

        e.votes.Store(voteID, vote)
    }

    return nil
}

// GetFeed returns a list of posts from subscribed subreddits
func (e *RedditEngine) GetFeed(userID string) ([]*models.Post, error) {
    var feed []*models.Post
    userSubscriptions := make(map[string]bool)

    // Get user's subscribed subreddits
    e.subreddits.Range(func(key, value interface{}) bool {
        subreddit := value.(*models.SubReddit)
        if _, isMember := subreddit.Members.Load(userID); isMember {
            userSubscriptions[subreddit.ID] = true
        }
        return true
    })

    // Collect posts from subscribed subreddits
    e.posts.Range(func(key, value interface{}) bool {
        post := value.(*models.Post)
        if userSubscriptions[post.SubRedditID] {
            feed = append(feed, post)
        }
        return true
    })

    return feed, nil
}

// SendDirectMessage sends a direct message from one user to another
func (e *RedditEngine) SendDirectMessage(fromID, toID, content string) (*models.DirectMessage, error) {
    // Validate both users exist
    _, fromExists := e.users.Load(fromID)
    _, toExists := e.users.Load(toID)

    if !fromExists {
        return nil, errors.New("sender not found")
    }
    if !toExists {
        return nil, errors.New("recipient not found")
    }

    message := &models.DirectMessage{
        ID:        generateID(),
        FromID:    fromID,
        ToID:      toID,
        Content:   content,
        CreatedAt: time.Now(),
    }

    e.messages.Store(message.ID, message)
    return message, nil
}

// GetMessage retrieves a single message
func (e *RedditEngine) GetMessage(userID, messageID string) (*models.DirectMessage, error) {
    msgI, ok := e.messages.Load(messageID)
    if !ok {
        return nil, errors.New("message not found")
    }
    msg := msgI.(*models.DirectMessage)
    // Check if user is either sender or recipient
    if msg.FromID != userID && msg.ToID != userID {
        return nil, errors.New("unauthorized access to message")
    }
    return msg, nil
}

func (e *RedditEngine) GetUserPublicKey(userID string) (string, error) {
    _, ok := e.users.Load(userID)  // Changed from userI, ok to _, ok
    if !ok {
        return "", errors.New("user not found")
    }
    return "dummy-public-key", nil
}

// GetUserMessages returns all messages for a user
func (e *RedditEngine) GetUserMessages(userID string) ([]*models.DirectMessage, error) {
    var messages []*models.DirectMessage
    e.messages.Range(func(_, value interface{}) bool {
        msg := value.(*models.DirectMessage)
        if msg.ToID == userID || msg.FromID == userID {
            messages = append(messages, msg)
        }
        return true
    })
    return messages, nil
}