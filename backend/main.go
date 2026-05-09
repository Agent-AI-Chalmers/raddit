package main

import (
	"log"
	"net/http"
	"os"
	"raddit/config"
	"raddit/database"
	"raddit/handlers"
	"raddit/middleware"
	"raddit/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize data store
	database.Initialize()

	// Ensure upload directory exists
	if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
		log.Fatal("Cannot create upload directory:", err)
	}

	if config.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Allow cross-origin requests from the frontend dev server.
	// AllowCredentials=true is required so browsers send the session cookie.
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Serve uploaded files statically
	r.Static("/uploads", config.UploadDir)

	api := r.Group("/api")
	{
		// Auth endpoints — open access
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/logout", handlers.Logout)
			auth.GET("/me", middleware.AuthRequired(), handlers.GetMe)
		}

		// Post endpoints
		posts := api.Group("/posts")
		{
			posts.GET("", handlers.ListPosts)
			posts.GET("/search", handlers.SearchPosts)
			posts.GET("/preview", handlers.PreviewURL)
			posts.GET("/:id", handlers.GetPost)
			posts.GET("/:id/comments", handlers.ListComments)
			posts.POST("", middleware.AuthRequired(), handlers.CreatePost)
			posts.PUT("/:id", middleware.AuthRequired(), handlers.UpdatePost)
			posts.DELETE("/:id", middleware.AuthRequired(), handlers.DeletePost)
			posts.POST("/:id/vote", middleware.AuthRequired(), handlers.VotePost)
		}

		// Comment endpoints
		comments := api.Group("/comments")
		{
			comments.POST("", middleware.AuthRequired(), handlers.CreateComment)
			comments.DELETE("/:id", middleware.AuthRequired(), handlers.DeleteComment)
		}

		// User endpoints
		users := api.Group("/users")
		{
			users.GET("/:id", handlers.GetUser)
			users.GET("/:id/posts", handlers.GetUserPosts)
			users.PUT("/profile", middleware.AuthRequired(), handlers.UpdateProfile)
		}

		// File management
		files := api.Group("/files")
		{
			files.POST("/upload", middleware.AuthRequired(), handlers.UploadFile)
			files.GET("/download", handlers.DownloadFile)
			files.GET("/list", middleware.AuthRequired(), handlers.ListFiles)
		}

		// Network utilities panel
		tools := api.Group("/tools")
		{
			tools.GET("/ping", utils.PingHost)
			tools.GET("/nslookup", utils.NslookupHost)
		}

		// Admin panel — requires authentication and admin role
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(), middleware.AdminRequired())
		{
			admin.GET("/dashboard", handlers.AdminDashboard)
			admin.GET("/debug", handlers.SystemInfo)
			admin.GET("/users", handlers.ListUsers)
			admin.DELETE("/users/:id", handlers.AdminDeleteUser)
		}
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "raddit"})
	})

	log.Printf("Raddit backend running on :%s", config.ServerPort)
	if err := r.Run(":" + config.ServerPort); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
