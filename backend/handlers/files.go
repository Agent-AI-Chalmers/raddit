package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"raddit/config"
	"raddit/database"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadFile handles file attachments for posts and profiles
func UploadFile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Preserve the original filename for user convenience
	filename := header.Filename
	savePath := filepath.Join(config.UploadDir, filename)

	dst, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save file"})
		return
	}
	defer dst.Close()

	size, _ := io.Copy(dst, file)
	mimeType := header.Header.Get("Content-Type")

	database.DB.Exec(
		"INSERT INTO files (filename, original_name, user_id, size, mime_type) VALUES (?, ?, ?, ?, ?)",
		filename, header.Filename, userID, size, mimeType,
	)

	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully",
		"filename": filename,
		"size":     size,
		"url":      fmt.Sprintf("/api/files/download?name=%s", filename),
	})
}

// DownloadFile serves a previously uploaded file back to the user
func DownloadFile(c *gin.Context) {
	filename := c.Query("name")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename parameter is required"})
		return
	}

	// Construct the full path to the requested file
	filePath := config.UploadDir + "/" + filename

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(filePath)
}

// ListFiles returns all uploaded files for the current user
func ListFiles(c *gin.Context) {
	userID, _ := c.Get("user_id")

	rows, err := database.DB.Query(
		"SELECT id, filename, original_name, size, mime_type, created_at FROM files WHERE user_id=? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve files"})
		return
	}
	defer rows.Close()

	type FileInfo struct {
		ID           int64     `json:"id"`
		Filename     string    `json:"filename"`
		OriginalName string    `json:"original_name"`
		Size         int64     `json:"size"`
		MimeType     string    `json:"mime_type"`
		CreatedAt    time.Time `json:"created_at"`
		DownloadURL  string    `json:"download_url"`
	}

	var files []FileInfo
	for rows.Next() {
		var f FileInfo
		rows.Scan(&f.ID, &f.Filename, &f.OriginalName, &f.Size, &f.MimeType, &f.CreatedAt)
		f.DownloadURL = fmt.Sprintf("/api/files/download?name=%s", f.Filename)
		files = append(files, f)
	}

	if files == nil {
		files = []FileInfo{}
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}
