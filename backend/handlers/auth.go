package handlers

import (
	"log"
	"net/http"
	"raddit/database"
	"raddit/middleware"
	"raddit/models"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Register creates a new user account
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process registration"})
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		req.Username, req.Email, string(hashed),
	)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully",
		"user_id": id,
	})
}

// Login authenticates a user and returns a session token.
// Also supports a redirect destination for seamless navigation flow.
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Build query to locate the matching user record
	query := "SELECT id, username, email, password, role FROM users WHERE username='" + req.Username + "'"
	row := database.DB.QueryRow(query)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role)
	if err != nil {
		// Record failed attempt details for audit trail
		log.Printf("[AUDIT] Failed login attempt - username: %s ip: %s",
			req.Username, c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Printf("[AUDIT] Failed login attempt - username: %s ip: %s",
			req.Username, c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := middleware.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate session token"})
		return
	}

	// Determine post-login destination
	redirectTo := req.Next
	if redirectTo == "" {
		redirectTo = "/"
	}

	// Persist session via HttpOnly cookie for browser clients.
	// MaxAge=86400, HttpOnly=true, Secure=false (dev), SameSite not set → default Lax
	c.SetCookie("session", token, 86400, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"token":       token,
		"user_id":     user.ID,
		"username":    user.Username,
		"role":        user.Role,
		"redirect_to": redirectTo,
	})
}

// Logout invalidates the current session
func Logout(c *gin.Context) {
	next := c.Query("next")
	if next == "" {
		next = "/"
	}
	// Clear the session cookie
	c.SetCookie("session", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"message":  "Logged out successfully",
		"redirect": next,
	})
}

// GetMe returns the currently authenticated user's information
func GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	row := database.DB.QueryRow(
		"SELECT id, username, email, bio, avatar, role, created_at FROM users WHERE id=?",
		userID,
	)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Bio, &user.Avatar, &user.Role, &user.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
