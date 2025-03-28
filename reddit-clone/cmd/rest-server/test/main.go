// cmd/rest-server/test/main.go
package main

import (
    "log"
    "time"
    
    "reddit-clone/internal/web"
)

func main() {
    // Create a new client
    client := web.NewClient("http://localhost:8080")

    // Test 1: Register and Login
    log.Println("=== Testing Registration and Login ===")
    err := client.Register("testuser", "password123")
    if err != nil {
        log.Printf("Registration failed: %v\n", err)
    } else {
        log.Println("Registration successful")
    }

    token, err := client.Login("testuser", "password123")
    if err != nil {
        log.Fatalf("Login failed: %v\n", err)
    }
    log.Printf("Login successful, token: %s\n", token)
    client.SetToken(token)

    // Test 2: Create Subreddit
    log.Println("\n=== Testing Subreddit Creation ===")
    subreddit, err := client.CreateSubreddit("TestSubreddit", "A test subreddit")
    if err != nil {
        log.Printf("Subreddit creation failed: %v\n", err)
    } else {
        log.Printf("Created subreddit: %s\n", subreddit.Name)
    }

    // Test 3: Create Post
    log.Println("\n=== Testing Post Creation ===")
    post, err := client.CreatePost(
        "Test Post",
        "This is a test post content",
        subreddit.ID,
    )
    if err != nil {
        log.Printf("Post creation failed: %v\n", err)
    } else {
        log.Printf("Created post: %s\n", post.Title)
    }

    // Test 4: Create Comment
    log.Println("\n=== Testing Comment Creation ===")
    _, err = client.CreateComment(
        "This is a test comment",
        post.ID,
        nil,
    )
    if err != nil {
        log.Printf("Comment creation failed: %v\n", err)
    } else {
        log.Printf("Created comment on post: %s\n", post.ID)
    }

    // Test 5: Vote
    log.Println("\n=== Testing Voting ===")
    err = client.Vote(post.ID, true) // Upvote
    if err != nil {
        log.Printf("Voting failed: %v\n", err)
    } else {
        log.Printf("Successfully voted on post: %s\n", post.ID)
    }

    // Test 6: Get Feed
    log.Println("\n=== Testing Feed Retrieval ===")
    feed, err := client.GetFeed()
    if err != nil {
        log.Printf("Feed retrieval failed: %v\n", err)
    } else {
        log.Printf("Retrieved %d posts from feed\n", len(feed))
        for _, p := range feed {
            log.Printf("- Post: %s\n", p.Title)
        }
    }

    // Wait a bit to see the results
    time.Sleep(time.Second)
    log.Println("\nTest completed")
}