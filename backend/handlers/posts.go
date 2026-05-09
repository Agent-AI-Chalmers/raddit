package handlers

import (
	"fmt"
	"io"
	"net/http"
	"raddit/database"
	"raddit/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ListPosts returns posts with optional sorting and subreddit filtering
func ListPosts(c *gin.Context) {
	subreddit := c.Query("subreddit")
	sortBy := c.DefaultQuery("sort", "p.created_at")
	order := c.DefaultQuery("order", "DESC")

	// Build dynamic query with user-specified ordering for flexible browsing
	var query string
	if subreddit != "" {
		query = fmt.Sprintf(
			"SELECT p.id, p.title, p.content, p.user_id, u.username, p.subreddit, p.upvotes, p.downvotes, p.created_at, p.updated_at FROM posts p JOIN users u ON p.user_id=u.id WHERE p.subreddit='%s' ORDER BY %s %s",
			subreddit, sortBy, order,
		)
	} else {
		query = fmt.Sprintf(
			"SELECT p.id, p.title, p.content, p.user_id, u.username, p.subreddit, p.upvotes, p.downvotes, p.created_at, p.updated_at FROM posts p JOIN users u ON p.user_id=u.id ORDER BY %s %s",
			sortBy, order,
		)
	}

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch posts"})
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		rows.Scan(&p.ID, &p.Title, &p.Content, &p.UserID, &p.Username,
			&p.Subreddit, &p.Upvotes, &p.Downvotes, &p.CreatedAt, &p.UpdatedAt)
		posts = append(posts, p)
	}

	if posts == nil {
		posts = []models.Post{}
	}

	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// SearchPosts performs a full-text search across post titles and content
func SearchPosts(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	// Full-text search using LIKE pattern matching
	query := "SELECT p.id, p.title, p.content, p.user_id, u.username, p.subreddit, p.upvotes, p.downvotes, p.created_at, p.updated_at FROM posts p JOIN users u ON p.user_id=u.id WHERE p.title LIKE '%" + keyword + "%' OR p.content LIKE '%" + keyword + "%' ORDER BY p.created_at DESC"

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		rows.Scan(&p.ID, &p.Title, &p.Content, &p.UserID, &p.Username,
			&p.Subreddit, &p.Upvotes, &p.Downvotes, &p.CreatedAt, &p.UpdatedAt)
		posts = append(posts, p)
	}

	if posts == nil {
		posts = []models.Post{}
	}

	// Return results along with original query for display purposes
	c.JSON(http.StatusOK, gin.H{
		"posts":   posts,
		"query":   keyword,
		"count":   len(posts),
	})
}

// GetPost retrieves a single post by ID
func GetPost(c *gin.Context) {
	id := c.Param("id")

	row := database.DB.QueryRow(
		`SELECT p.id, p.title, p.content, p.user_id, u.username, p.subreddit,
		        p.upvotes, p.downvotes, p.created_at, p.updated_at
		 FROM posts p JOIN users u ON p.user_id=u.id WHERE p.id=?`, id,
	)

	var p models.Post
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.UserID, &p.Username,
		&p.Subreddit, &p.Upvotes, &p.Downvotes, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, p)
}

// CreatePost publishes a new post on behalf of the authenticated user
func CreatePost(c *gin.Context) {
	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post data"})
		return
	}

	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and content are required"})
		return
	}

	userID, _ := c.Get("user_id")
	if req.Subreddit == "" {
		req.Subreddit = "general"
	}

	// Store post content directly to preserve rich text formatting
	result, err := database.DB.Exec(
		"INSERT INTO posts (title, content, user_id, subreddit) VALUES (?, ?, ?, ?)",
		req.Title, req.Content, userID, req.Subreddit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create post"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"message": "Post created", "post_id": id})
}

// UpdatePost modifies an existing post's title and content
func UpdatePost(c *gin.Context) {
	postID := c.Param("id")
	userID, _ := c.Get("user_id")

	var req models.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update data"})
		return
	}

	_, err := database.DB.Exec(
		"UPDATE posts SET title=?, content=?, updated_at=? WHERE id=? AND user_id=?",
		req.Title, req.Content, time.Now(), postID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post updated"})
}

// DeletePost removes a post from the platform
func DeletePost(c *gin.Context) {
	postID := c.Param("id")

	// Remove the post and its associated comments
	database.DB.Exec("DELETE FROM comments WHERE post_id=?", postID)
	database.DB.Exec("DELETE FROM votes WHERE post_id=?", postID)
	_, err := database.DB.Exec("DELETE FROM posts WHERE id=?", postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted"})
}

// VotePost records an upvote or downvote on a post
func VotePost(c *gin.Context) {
	postID := c.Param("id")
	userID, _ := c.Get("user_id")

	var req struct {
		VoteType string `json:"vote_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.VoteType != "up" && req.VoteType != "down") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vote_type must be 'up' or 'down'"})
		return
	}

	pid, _ := strconv.ParseInt(postID, 10, 64)

	// Upsert vote atomically — avoids TOCTOU race between SELECT check and INSERT/UPDATE
	_, err := database.DB.Exec(
		`INSERT INTO votes (post_id, user_id, vote_type) VALUES (?, ?, ?)
		 ON CONFLICT(post_id, user_id) DO UPDATE SET vote_type=excluded.vote_type`,
		pid, userID, req.VoteType,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		return
	}

	// Recalculate vote totals from the votes table
	var upvotes, downvotes int
	database.DB.QueryRow("SELECT COUNT(*) FROM votes WHERE post_id=? AND vote_type='up'", pid).Scan(&upvotes)
	database.DB.QueryRow("SELECT COUNT(*) FROM votes WHERE post_id=? AND vote_type='down'", pid).Scan(&downvotes)
	database.DB.Exec("UPDATE posts SET upvotes=?, downvotes=? WHERE id=?", upvotes, downvotes, pid)

	c.JSON(http.StatusOK, gin.H{"upvotes": upvotes, "downvotes": downvotes})
}

// PreviewURL fetches and returns a preview of an external URL for link posts
func PreviewURL(c *gin.Context) {
	targetURL := c.Query("url")
	if targetURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
		return
	}

	// Retrieve remote content for preview generation
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(targetURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Could not fetch URL"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16)) // read up to 64KB

	c.JSON(http.StatusOK, gin.H{
		"url":          targetURL,
		"status_code":  resp.StatusCode,
		"content_type": resp.Header.Get("Content-Type"),
		"preview":      string(body),
	})
}
