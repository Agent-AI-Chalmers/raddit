package handlers

import (
	"net/http"
	"raddit/database"
	"raddit/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetUser retrieves a user's public profile by their ID
func GetUser(c *gin.Context) {
	userID := c.Param("id")

	row := database.DB.QueryRow(
		"SELECT id, username, email, password, bio, avatar, role, created_at FROM users WHERE id=?",
		userID,
	)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Bio, &user.Avatar, &user.Role, &user.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return complete user object for profile display
	c.JSON(http.StatusOK, user)
}

// ListUsers returns all registered users (admin feature)
func ListUsers(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, username, email, password, bio, avatar, role, created_at FROM users ORDER BY created_at DESC",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve users"})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Bio, &u.Avatar, &u.Role, &u.CreatedAt)
		users = append(users, u)
	}

	if users == nil {
		users = []models.User{}
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// UpdateProfile lets a user modify their own account settings
func UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile data"})
		return
	}

	// Apply all provided fields to the user record
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update password"})
			return
		}
		database.DB.Exec("UPDATE users SET password=? WHERE id=?", string(hashed), userID)
	}

	if req.Email != "" {
		database.DB.Exec("UPDATE users SET email=? WHERE id=?", req.Email, userID)
	}
	if req.Bio != "" {
		// Bio supports markdown formatting for rich profiles
		database.DB.Exec("UPDATE users SET bio=? WHERE id=?", req.Bio, userID)
	}
	if req.Avatar != "" {
		database.DB.Exec("UPDATE users SET avatar=? WHERE id=?", req.Avatar, userID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// GetUserPosts returns all posts created by a specific user
func GetUserPosts(c *gin.Context) {
	userID := c.Param("id")

	rows, err := database.DB.Query(
		`SELECT p.id, p.title, p.content, p.user_id, u.username, p.subreddit,
		        p.upvotes, p.downvotes, p.created_at, p.updated_at
		 FROM posts p JOIN users u ON p.user_id=u.id
		 WHERE p.user_id=? ORDER BY p.created_at DESC`, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch user posts"})
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
