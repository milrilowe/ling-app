package main

import (
	"context"
	"log"
	"os"
	"time"

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

	// Initialize storage service
	storageService, err := services.NewStorageService(
		cfg.S3Endpoint,
		cfg.S3AccessKey,
		cfg.S3SecretKey,
		cfg.S3Bucket,
		cfg.S3Region,
	)
	if err != nil {
		log.Fatal("Failed to initialize storage service:", err)
		os.Exit(1)
	}

	// Ensure MinIO bucket exists
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := storageService.EnsureBucketExists(ctx); err != nil {
		log.Printf("Warning: Failed to ensure bucket exists: %v", err)
	} else {
		log.Printf("Storage bucket '%s' is ready", cfg.S3Bucket)
	}

	// Initialize AI clients
	openAIClient := services.NewOpenAIClient(cfg.OpenAIAPIKey)
	whisperClient := services.NewWhisperClient(cfg.OpenAIAPIKey)
	elevenLabsClient := services.NewElevenLabsClient(cfg.ElevenLabsAPIKey)

	// Initialize ML client for pronunciation analysis
	mlClient := services.NewMLClient(cfg.MLServiceURL, time.Duration(cfg.MLServiceTimeout)*time.Second)

	// Initialize pronunciation worker
	pronunciationWorker := services.NewPronunciationWorker(database, mlClient, storageService)

	// Initialize handlers
	threadHandler := handlers.NewThreadHandler(database, openAIClient, storageService, whisperClient, elevenLabsClient, pronunciationWorker)
	audioHandler := handlers.NewAudioHandler(storageService)

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
		api.POST("/threads/:id/messages/audio", threadHandler.SendAudioMessage)

		// Audio - use *key to capture full path including slashes
		api.GET("/audio/*key", audioHandler.GetAudio)
	}

	// Start server
	addr := cfg.Host + ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
		os.Exit(1)
	}
}
