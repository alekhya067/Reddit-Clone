// internal/simulator/simulator.go
package simulator

import (
    "fmt"
    "log"
    "math/rand"
    "sync"
    "time"
    
    "reddit-clone/internal/client"
    "reddit-clone/internal/models"
)

type Simulator struct {
    client         *client.RedditClient
    users          []*models.User
    subreddits     []*models.SubReddit
    numUsers       int
    activeUsers    sync.Map            // map[string]bool
    userSubs       map[string][]string // map[userID][]subredditID
    subredditNames map[string]string   // map[subredditID]name
    postCount      map[string]int      // map[subredditID]count
    commentCount   map[string]int      // map[subredditID]count
    voteCount      map[string]int      // map[subredditID]count
    rng            *rand.Rand
    wg             sync.WaitGroup
    stopChan       chan struct{}
    metrics        *models.Metrics
    mtx            sync.RWMutex
}

func NewSimulator(client *client.RedditClient, numUsers int) *Simulator {
    return &Simulator{
        client:         client,
        numUsers:       numUsers,
        userSubs:       make(map[string][]string),
        subredditNames: make(map[string]string),
        postCount:      make(map[string]int),
        commentCount:   make(map[string]int),
        voteCount:      make(map[string]int),
        rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
        stopChan:      make(chan struct{}),
        metrics:       &models.Metrics{
            StartTime:      time.Now(),
            SubredditStats: make(map[string]*models.SubredditMetrics),
        },
    }
}

func (s *Simulator) Start() {
    log.Printf("Starting simulation with %d users...\n", s.numUsers)
    
    // Initialize users and subreddits
    s.initializeEnvironment()
    
    // Start user simulations
    s.simulateUsers()
}

func (s *Simulator) Stop() {
    close(s.stopChan)
    s.wg.Wait()
    log.Println("Simulation stopped")
}

func (s *Simulator) initializeEnvironment() {
    // Create users
    for i := 0; i < s.numUsers; i++ {
        username := fmt.Sprintf("user_%d", i)
        user, err := s.client.RegisterAccount(username, "password123")
        if err != nil {
            log.Printf("Error creating user %s: %v\n", username, err)
            continue
        }
        s.users = append(s.users, user)
        s.userSubs[user.ID] = make([]string, 0)
    }
    if len(s.users) == 0 {
        log.Fatal("No users were created successfully")
    }

    // Create subreddits (20% of user count, minimum 5)
    numSubreddits := max(5, s.numUsers/5)
    log.Printf("Creating %d subreddits...\n", numSubreddits)
    
    for i := 0; i < numSubreddits; i++ {
        name := fmt.Sprintf("subreddit_%d", i)
        creatorIndex := i % len(s.users)
        subreddit, err := s.client.CreateSubReddit(
            name,
            fmt.Sprintf("Description for %s", name),
            s.users[creatorIndex].ID,
        )
        if err != nil {
            log.Printf("Error creating subreddit %s: %v\n", name, err)
            continue
        }
        s.subreddits = append(s.subreddits, subreddit)
        s.userSubs[s.users[creatorIndex].ID] = append(
            s.userSubs[s.users[creatorIndex].ID],
            subreddit.ID,
        )
        s.subredditNames[subreddit.ID] = name
    }

    // Simulate Zipf distribution for subreddit memberships
    zipf := s.generateZipfDistribution()
    
    for _, user := range s.users {
        // Each user joins 2-5 subreddits
        numToJoin := 2 + int(zipf.Uint64())%4
        joinedSubs := make(map[string]bool)
        
        for j := 0; j < numToJoin; j++ {
            subredditIndex := int(zipf.Uint64()) % len(s.subreddits)
            subreddit := s.subreddits[subredditIndex]
            
            if !joinedSubs[subreddit.ID] {
                err := s.client.JoinSubReddit(user.ID, subreddit.ID)
                if err != nil {
                    log.Printf("Error joining subreddit: %v\n", err)
                    continue
                }
                s.userSubs[user.ID] = append(s.userSubs[user.ID], subreddit.ID)
                joinedSubs[subreddit.ID] = true
                
                log.Printf("User %s joined %s\n", user.Username, s.subredditNames[subreddit.ID])
            }
        }
    }
}

func (s *Simulator) simulateUsers() {
    for _, user := range s.users {
        s.wg.Add(1)
        go func(u *models.User) {
            defer s.wg.Done()
            s.simulateUserActivity(u)
        }(user)
    }
}

func (s *Simulator) simulateUserActivity(user *models.User) {
    ticker := time.NewTicker(time.Duration(1+s.rng.Intn(4)) * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-s.stopChan:
            return
            
        case <-ticker.C:
            // Simulate connection/disconnection
            isActive := s.rng.Float64() < 0.7 // 70% chance of being active
            s.activeUsers.Store(user.ID, isActive)

            if !isActive {
                continue
            }

            // Perform random actions
            switch s.rng.Intn(5) {
            case 0:
                s.simulatePosting(user)
            case 1:
                s.simulateCommenting(user)
            case 2:
                s.simulateVoting(user)
            case 3:
                s.simulateRepost(user)
            case 4:
                s.simulateDirectMessage(user)
            }
        }
    }
}

// Continue with the action simulation methods...

// Continuing simulator.go...

// In simulatePosting method in internal/simulator/simulator.go

func (s *Simulator) simulatePosting(user *models.User) {
    userSubs := s.userSubs[user.ID]
    if len(userSubs) == 0 {
        return
    }

    // Select a random subreddit to post in
    subID := userSubs[s.rng.Intn(len(userSubs))]
    
    // Remove the unused variable declaration
    _, err := s.client.CreatePost(
        fmt.Sprintf("Post by %s in %s", user.Username, s.subredditNames[subID]),
        fmt.Sprintf("Content from user %s at %s", user.Username, time.Now().Format(time.RFC3339)),
        user.ID,
        subID,
    )
    
    if err != nil {
        log.Printf("Error creating post: %v\n", err)
        return
    }

    s.mtx.Lock()
    s.postCount[subID]++
    s.metrics.TotalPosts++
    s.mtx.Unlock()

    log.Printf("User %s created post in %s\n", user.Username, s.subredditNames[subID])
}
func (s *Simulator) simulateCommenting(user *models.User) {
    feed, err := s.client.GetFeed(user.ID)
    if err != nil || len(feed) == 0 {
        return
    }

    // Select a random post to comment on
    post := feed[s.rng.Intn(len(feed))]
    
    comment, err := s.client.CreateComment(
        fmt.Sprintf("Comment from %s at %s", user.Username, time.Now().Format(time.RFC3339)),
        user.ID,
        post.ID,
        nil, // Top-level comment
    )
    
    if err != nil {
        log.Printf("Error creating comment: %v\n", err)
        return
    }

    s.mtx.Lock()
    s.commentCount[post.SubRedditID]++
    s.metrics.TotalComments++
    s.mtx.Unlock()

    // 30% chance to create a nested comment
    if s.rng.Float64() < 0.3 {
        _, err = s.client.CreateComment(
            fmt.Sprintf("Nested comment from %s", user.Username),
            user.ID,
            post.ID,
            &comment.ID,
        )
        if err != nil {
            log.Printf("Error creating nested comment: %v\n", err)
        }
    }
}

func (s *Simulator) simulateVoting(user *models.User) {
    feed, err := s.client.GetFeed(user.ID)
    if err != nil || len(feed) == 0 {
        return
    }

    post := feed[s.rng.Intn(len(feed))]
    isUpvote := s.rng.Float64() < 0.7 // 70% chance of upvote
    
    err = s.client.Vote(user.ID, post.ID, isUpvote)
    if err != nil {
        log.Printf("Error voting: %v\n", err)
        return
    }

    s.mtx.Lock()
    s.voteCount[post.SubRedditID]++
    s.metrics.TotalVotes++
    s.mtx.Unlock()
}

func (s *Simulator) simulateRepost(user *models.User) {
    feed, err := s.client.GetFeed(user.ID)
    if err != nil || len(feed) == 0 {
        return
    }

    // Find a popular post to repost
    var popularPosts []*models.Post
    for _, post := range feed {
        if post.Upvotes > 10 && post.AuthorID != user.ID {
            popularPosts = append(popularPosts, post)
        }
    }

    if len(popularPosts) == 0 {
        return
    }

    originalPost := popularPosts[s.rng.Intn(len(popularPosts))]
    userSubs := s.userSubs[user.ID]
    if len(userSubs) == 0 {
        return
    }

    // Repost to a random subreddit the user is subscribed to
    targetSubID := userSubs[s.rng.Intn(len(userSubs))]
    
    _, err = s.client.CreatePost(
        fmt.Sprintf("[Repost] %s", originalPost.Title),
        originalPost.Content,
        user.ID,
        targetSubID,
    )
    
    if err != nil {
        log.Printf("Error creating repost: %v\n", err)
    }
}

func (s *Simulator) simulateDirectMessage(user *models.User) {
    if len(s.users) <= 1 {
        return
    }

    // Select a random recipient that's not the sender
    var recipient *models.User
    for {
        recipient = s.users[s.rng.Intn(len(s.users))]
        if recipient.ID != user.ID {
            break
        }
    }

    _, err := s.client.SendDirectMessage(
        user.ID,
        recipient.ID,
        fmt.Sprintf("Message from %s at %s", user.Username, time.Now().Format(time.RFC3339)),
    )
    
    if err != nil {
        log.Printf("Error sending message: %v\n", err)
    }
}

// Helper methods
func (s *Simulator) generateZipfDistribution() *rand.Zipf {
    return rand.NewZipf(s.rng, 1.5, 1, uint64(max(1, len(s.subreddits))))
}

func (s *Simulator) GetMetrics() *models.Metrics {
    s.mtx.RLock()
    defer s.mtx.RUnlock()

    // Count active users
    var activeCount int64
    s.activeUsers.Range(func(_, value interface{}) bool {
        if value.(bool) {
            activeCount++
        }
        return true
    })

    s.metrics.ActiveUsers = activeCount
    
    return s.metrics
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}