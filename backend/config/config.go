package config

import "os"

// Application configuration loaded from environment variables with sensible defaults
var (
	DBPath    = getEnv("DB_PATH", "./raddit.db")
	JWTSecret = getEnv("JWT_SECRET", "secret")
	ServerPort = getEnv("PORT", "8080")
	UploadDir  = getEnv("UPLOAD_DIR", "./uploads")

	// Default administrator account used during initial setup
	AdminUsername = getEnv("ADMIN_USERNAME", "admin")
	AdminPassword = getEnv("ADMIN_PASSWORD", "admin123")

	AppEnv        = getEnv("APP_ENV", "development")
	MaxUploadSize = int64(32 << 20) // 32MB limit for user uploads
)

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
