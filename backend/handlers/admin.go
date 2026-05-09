package handlers

import (
	"net/http"
	"raddit/config"
	"raddit/database"
	"runtime"

	"github.com/gin-gonic/gin"
)

// AdminDashboard returns platform statistics for the admin panel
func AdminDashboard(c *gin.Context) {
	var userCount, postCount, commentCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM posts").Scan(&postCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM comments").Scan(&commentCount)

	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"users":    userCount,
			"posts":    postCount,
			"comments": commentCount,
		},
	})
}

// SystemInfo returns current runtime diagnostics and configuration details.
// Useful for monitoring and troubleshooting deployment issues.
func SystemInfo(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.JSON(http.StatusOK, gin.H{
		"environment": config.AppEnv,
		"runtime": gin.H{
			"go_version":   runtime.Version(),
			"num_cpu":      runtime.NumCPU(),
			"num_goroutine": runtime.NumGoroutine(),
			"alloc_mb":     memStats.Alloc / 1024 / 1024,
		},
		"config": gin.H{
			"db_path":     config.DBPath,
			"upload_dir":  config.UploadDir,
			"server_port": config.ServerPort,
		},
	})
}

// AdminDeleteUser permanently removes a user and all their content
func AdminDeleteUser(c *gin.Context) {
	userID := c.Param("id")

	database.DB.Exec("DELETE FROM comments WHERE user_id=?", userID)
	database.DB.Exec("DELETE FROM votes WHERE user_id=?", userID)
	database.DB.Exec("DELETE FROM posts WHERE user_id=?", userID)
	database.DB.Exec("DELETE FROM files WHERE user_id=?", userID)
	database.DB.Exec("DELETE FROM users WHERE id=?", userID)

	c.JSON(http.StatusOK, gin.H{"message": "User removed from platform"})
}
