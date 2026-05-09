package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"raddit/config"
	"raddit/database"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// allowedExtensions defines file extensions that are safe to serve statically.
// Executable, script, and markup types are excluded to prevent stored XSS.
var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".svg":  false, // SVG can contain embedded scripts
	".pdf":  true,
	".txt":  true,
	".csv":  true,
	".mp3":  true,
	".mp4":  true,
	".webm": true,
	".ogg":  true,
	".wav":  true,
	".zip":  true,
	".gz":   true,
	".tar":  true,
}

// allowedMIMETypes defines content types that are safe to store and serve.
var allowedMIMETypes = map[string]bool{
	"image/jpeg":        true,
	"image/png":         true,
	"image/gif":         true,
	"image/webp":        true,
	"application/pdf":   true,
	"text/plain":        true,
	"text/csv":          true,
	"audio/mpeg":        true,
	"audio/ogg":         true,
	"audio/wav":         true,
	"video/mp4":         true,
	"video/webm":        true,
	"application/zip":   true,
	"application/gzip":  true,
	"application/x-tar": true,
}

// validateFileExtension checks that the file extension is in the allowlist.
func validateFileExtension(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return "", fmt.Errorf("file must have an extension")
	}
	allowed, ok := allowedExtensions[ext]
	if !ok || !allowed {
		return "", fmt.Errorf("file type %s is not allowed", ext)
	}
	return ext, nil
}

// detectContentType reads the first 512 bytes to determine the actual content type,
// then verifies it against the allowlist.
func detectContentType(file multipart.File) (string, error) {
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("could not read file for type detection")
	}
	buf = buf[:n]

	// Reset read position for subsequent copy
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("could not reset file position")
	}

	detected := http.DetectContentType(buf)
	// Strip charset/params from detected type (e.g. "text/plain; charset=utf-8" -> "text/plain")
	if idx := strings.IndexByte(detected, ';'); idx != -1 {
		detected = strings.TrimSpace(detected[:idx])
	}

	if !allowedMIMETypes[detected] {
		return "", fmt.Errorf("content type %s is not allowed", detected)
	}
	return detected, nil
}

// sanitizeFilename removes path separators and dangerous characters from the filename.
func sanitizeFilename(name string) string {
	// Strip any directory components
	name = filepath.Base(name)
	// Remove null bytes
	name = strings.ReplaceAll(name, "\x00", "")
	return name
}

// generateUniqueFilename creates a random hex name while preserving the extension.
func generateUniqueFilename(ext string) string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b) + ext
}

// UploadFile handles file attachments for posts and profiles
func UploadFile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// Enforce maximum upload size at the request level
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, config.MaxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Sanitize and validate the original filename extension
	originalName := sanitizeFilename(header.Filename)
	if originalName == "." || originalName == "/" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filename"})
		return
	}
	ext, err := validateFileExtension(originalName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Detect actual content type from file bytes (not client-supplied header)
	mimeType, err := detectContentType(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a unique storage filename to prevent overwrites and path traversal
	filename := generateUniqueFilename(ext)
	savePath := filepath.Join(config.UploadDir, filename)

	dst, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save file"})
		return
	}
	defer dst.Close()

	// Defense-in-depth: limit copy size even though MaxBytesReader is in place
	limitedReader := io.LimitReader(file, config.MaxUploadSize+1)
	size, _ := io.Copy(dst, limitedReader)
	if size > config.MaxUploadSize {
		dst.Close()
		os.Remove(savePath)
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": fmt.Sprintf("File exceeds maximum upload size of %d bytes", config.MaxUploadSize),
		})
		return
	}

	database.DB.Exec(
		"INSERT INTO files (filename, original_name, user_id, size, mime_type) VALUES (?, ?, ?, ?, ?)",
		filename, originalName, userID, size, mimeType,
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

	// Sanitize: strip directory components to prevent path traversal
	filename = filepath.Base(filename)
	if filename == "." || filename == string(os.PathSeparator) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filename"})
		return
	}

	userID, _ := c.Get("user_id")

	// Verify the file belongs to the requesting user (authorization check)
	var dbFilename string
	err := database.DB.QueryRow(
		"SELECT filename FROM files WHERE filename=? AND user_id=?",
		filename, userID,
	).Scan(&dbFilename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Construct the full path using filepath.Join (not string concatenation)
	filePath := filepath.Join(config.UploadDir, dbFilename)

	// Containment check: resolved path must stay within the upload directory
	absUploadDir, err := filepath.Abs(config.UploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server configuration error"})
		return
	}
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filename"})
		return
	}
	if !strings.HasPrefix(absFilePath, absUploadDir+string(os.PathSeparator)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(absFilePath)
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
