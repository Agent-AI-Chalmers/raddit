package handlers

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
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

	// Check for existing vote
	var existingType string
	err := database.DB.QueryRow(
		"SELECT vote_type FROM votes WHERE post_id=? AND user_id=?", pid, userID,
	).Scan(&existingType)

	if err == nil {
		// Update existing vote
		database.DB.Exec("UPDATE votes SET vote_type=? WHERE post_id=? AND user_id=?",
			req.VoteType, pid, userID)
	} else {
		// Insert new vote — note: no row-level locking applied here
		database.DB.Exec("INSERT INTO votes (post_id, user_id, vote_type) VALUES (?, ?, ?)",
			pid, userID, req.VoteType)
	}

	// Recalculate vote totals from the votes table
	var upvotes, downvotes int
	database.DB.QueryRow("SELECT COUNT(*) FROM votes WHERE post_id=? AND vote_type='up'", pid).Scan(&upvotes)
	database.DB.QueryRow("SELECT COUNT(*) FROM votes WHERE post_id=? AND vote_type='down'", pid).Scan(&downvotes)
	database.DB.Exec("UPDATE posts SET upvotes=?, downvotes=? WHERE id=?", upvotes, downvotes, pid)

	c.JSON(http.StatusOK, gin.H{"upvotes": upvotes, "downvotes": downvotes})
}

// isPrivateOrReservedIP checks whether an IP address belongs to a private,
// reserved, loopback, link-local, multicast, or unspecified range.
func isPrivateOrReservedIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified()
}

// validatePreviewURL checks that a URL is safe to fetch: only http/https schemes
// are allowed, the hostname must not be an internal/literal address, and DNS
// resolution must not yield a private or reserved IP. Returns an error if the
// URL is rejected.
func validatePreviewURL(ctx context.Context, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("only http and https URLs are allowed")
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return fmt.Errorf("URL has no hostname")
	}

	// Block obvious internal hostnames
	lowerHost := strings.ToLower(hostname)
	blockedHosts := []string{
		"localhost",
		"metadata.google.internal",
		"169.254.169.254",
	}
	for _, blocked := range blockedHosts {
		if lowerHost == blocked {
			return fmt.Errorf("access to internal host %q is not allowed", hostname)
		}
	}

	// Resolve hostname and verify the IP is not in a private/reserved range.
	// This also catches literal private IPs passed as hostnames.
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		return fmt.Errorf("could not resolve hostname: %w", err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("hostname resolved to no addresses")
	}
	for _, addr := range ips {
		if isPrivateOrReservedIP(addr.IP) {
			return fmt.Errorf("access to internal/private address is not allowed")
		}
	}

	return nil
}

// PreviewURL fetches and returns a preview of an external URL for link posts
func PreviewURL(c *gin.Context) {
	targetURL := c.Query("url")
	if targetURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
		return
	}

	ctx := c.Request.Context()

	// Validate the initial URL: scheme, hostname blocklist, and DNS → IP range check
	if err := validatePreviewURL(ctx, targetURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL not allowed: " + err.Error()})
		return
	}

	// Custom transport with DialContext that validates the resolved IP at
	// connection time, closing the DNS-rebinding TOCTOU window.
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, fmt.Errorf("invalid address: %w", err)
			}
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, fmt.Errorf("DNS resolution failed: %w", err)
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("no addresses for host %s", host)
			}
			for _, addr := range ips {
				if isPrivateOrReservedIP(addr.IP) {
					return nil, fmt.Errorf("connection to private/internal IP %s is not allowed", addr.IP)
				}
			}
			// Use the first validated IP for the actual connection
			return net.Dial(network, net.JoinHostPort(ips[0].IP.String(), port))
		},
	}

	// Custom redirect policy: re-validate every redirect target so that an
	// attacker cannot bypass IP checks by redirecting to an internal address.
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			if err := validatePreviewURL(req.Context(), req.URL.String()); err != nil {
				return fmt.Errorf("redirect target not allowed: %w", err)
			}
			return nil
		},
	}

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
