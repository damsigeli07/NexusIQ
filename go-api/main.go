package main

import (
	"log"

	"github.com/damsigeli07/NexusIQ/internal/config"
	"github.com/damsigeli07/NexusIQ/internal/db"
	"github.com/damsigeli07/NexusIQ/internal/handlers"
	"github.com/damsigeli07/NexusIQ/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Connect to Postgres and Redis
	dbConn := db.Connect(cfg)
	defer dbConn.Close()

	redisClient := db.ConnectRedis(cfg)
	defer redisClient.Close()

	r := gin.Default()

	// Health check — always useful for Docker health probes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "go-api"})
	})

	// Add this:
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "go-api"})
	})

	// Public routes
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", handlers.Register(dbConn))
		auth.POST("/login", handlers.Login(dbConn, cfg))
	}

	// Protected routes — JWT required
	api := r.Group("/api")
	api.Use(middleware.JWTAuth(cfg))
	{
		api.GET("/documents", handlers.ListDocuments(dbConn))
		api.POST("/documents", handlers.UploadDocument(dbConn, cfg))
		api.DELETE("/documents/:id", handlers.DeleteDocument(dbConn))

		api.GET("/chat", handlers.ChatWebSocket(dbConn, redisClient, cfg))

		api.GET("/analytics", handlers.GetAnalytics(dbConn))
	}

	log.Printf("Go API running on :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
