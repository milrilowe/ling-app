package main

import (
	"log"
	"os"

	"ling-app/api/internal/config"
	"ling-app/api/internal/db"
	"ling-app/api/internal/handlers"
	"ling-app/api/internal/middleware"
	"ling-app/api/internal/models"
	"ling-app/api/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		os.Exit(1)
	}

	// Run migrations
	if err := database.RunMigrations(&models.Thread{}, &models.Message{}); err != nil {
		log.Fatal("Failed to run migrations:", err)
		os.Exit(1)
	}

	// Initialize OpenAI client
	openAIClient := services.NewOpenAIClient(cfg.OpenAIAPIKey)

	// Initialize handlers
	threadHandler := handlers.NewThreadHandler(database, openAIClient)

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)

	// API routes
	api := router.Group("/api")
	{
		// Prompts
		api.GET("/prompts/random", handlers.GetRandomPrompt)

		// Threads
		api.GET("/threads", threadHandler.GetThreads)
		api.POST("/threads", threadHandler.CreateThread)
		api.GET("/threads/:id", threadHandler.GetThread)
		api.POST("/threads/:id/messages", threadHandler.SendMessage)
	}

	// Start server
	addr := cfg.Host + ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
		os.Exit(1)
	}
}
