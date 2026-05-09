package config

import (
	"log"
	"os"
)

// Application configuration loaded from environment variables.
// Sensitive credentials (JWT_SECRET, ADMIN_USERNAME, ADMIN_PASSWORD) are
// required and will cause startup failure if not provided.
var (
	DBPath     = getEnv("DB_PATH", "./raddit.db")
	JWTSecret  = getRequiredEnv("JWT_SECRET")
	ServerPort = getEnv("PORT", "8080")
	UploadDir  = getEnv("UPLOAD_DIR", "./uploads")

	// Administrator account credentials — must be supplied via environment
	AdminUsername = getRequiredEnv("ADMIN_USERNAME")
	AdminPassword = getRequiredEnv("ADMIN_PASSWORD")

	AppEnv        = getEnv("APP_ENV", "development")
	MaxUploadSize = int64(32 << 20) // 32MB limit for user uploads
)

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getRequiredEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return val
}
