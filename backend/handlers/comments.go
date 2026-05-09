package handlers

import (
	"net/http"
	"raddit/database"
	"raddit/models"
	"raddit/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListComments returns all comments for a specific post
func ListComments(c *gin.Context) {
	postID := c.Param("id")

	rows, err := database.DB.Query(
		`SELECT c.id, c.content, c.post_id, c.user_id, u.username, c.parent_id, c.created_at
		 FROM comments c JOIN users u ON c.user_id=u.id
		 WHERE c.post_id=? ORDER BY c.created_at ASC`, postID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch comments"})
		return
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var cm models.Comment
		rows.Scan(&cm.ID, &cm.Content, &cm.PostID, &cm.UserID, &cm.Username,
			&cm.ParentID, &cm.CreatedAt)
		comments = append(comments, cm)
	}

	if comments == nil {
		comments = []models.Comment{}
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

// CreateComment adds a reply to a post or another comment
func CreateComment(c *gin.Context) {
	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment data"})
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment content cannot be empty"})
		return
	}

	userID, _ := c.Get("user_id")

	// Sanitize user-supplied HTML to prevent stored XSS
	sanitizedContent := utils.SanitizeHTML(req.Content)

	result, err := database.DB.Exec(
		"INSERT INTO comments (content, post_id, user_id, parent_id) VALUES (?, ?, ?, ?)",
		sanitizedContent, req.PostID, userID, req.ParentID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not post comment"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"message": "Comment posted", "comment_id": id})
}

// DeleteComment removes a comment by ID
func DeleteComment(c *gin.Context) {
	commentID := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var ownerID int64
	err := database.DB.QueryRow("SELECT user_id FROM comments WHERE id=?", commentID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	// Allow deletion if user owns the comment or is an admin
	if ownerID != userID.(int64) && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete another user's comment"})
		return
	}

	database.DB.Exec("DELETE FROM comments WHERE id=?", commentID)
	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}
