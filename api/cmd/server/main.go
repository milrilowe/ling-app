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

	// Run migrations (order matters for foreign keys)
	if err := database.RunMigrations(
		&models.User{},
		&models.Session{},
		&models.Thread{},
		&models.Message{},
		&models.Subscription{},
		&models.Credits{},
		&models.CreditTransaction{},
		&models.PhonemeStats{},
		&models.PhonemeSubstitution{},
	); err != nil {
		log.Fatal("Failed to run migrations:", err)
		os.Exit(1)
	}

	// Initialize auth services
	authService := services.NewAuthService(database, cfg.SessionMaxAge)
	oauthService := services.NewOAuthService(cfg)

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
	whisperClient := services.NewWhisperClient(cfg.MLServiceURL) // Uses local faster-whisper via ML service
	ttsClient := services.NewTTSClient(cfg.MLServiceURL)         // Uses local Chatterbox TTS via ML service

	// Initialize ML client for pronunciation analysis
	mlClient := services.NewMLClient(cfg.MLServiceURL, time.Duration(cfg.MLServiceTimeout)*time.Second)

	// Initialize phoneme stats service
	phonemeStatsService := services.NewPhonemeStatsService(database)

	// Initialize pronunciation worker
	pronunciationWorker := services.NewPronunciationWorker(database, mlClient, storageService, phonemeStatsService)

	// Initialize credits and subscription services
	creditsService := services.NewCreditsService(database)
	stripeService := services.NewStripeService(cfg, database, creditsService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, oauthService, cfg)
	threadHandler := handlers.NewThreadHandler(database, openAIClient, storageService, whisperClient, ttsClient, pronunciationWorker, creditsService, cfg.MaxAudioFileSize)
	audioHandler := handlers.NewAudioHandler(storageService)
	subscriptionHandler := handlers.NewSubscriptionHandler(stripeService, creditsService)
	phonemeStatsHandler := handlers.NewPhonemeStatsHandler(phonemeStatsService)

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
		// Public routes (no auth required)
		api.GET("/prompts/random", handlers.GetRandomPrompt)

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			// /me requires authentication
			auth.GET("/me", middleware.RequireAuth(authService), authHandler.GetMe)
			// OAuth routes
			auth.GET("/google", authHandler.GoogleLogin)
			auth.GET("/google/callback", authHandler.GoogleCallback)
			auth.GET("/github", authHandler.GitHubLogin)
			auth.GET("/github/callback", authHandler.GitHubCallback)
		}

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(middleware.RequireAuth(authService))
		{
			// Threads
			protected.GET("/threads", threadHandler.GetThreads)
			protected.GET("/threads/archived", threadHandler.GetArchivedThreads)
			protected.POST("/threads", threadHandler.CreateThread)
			protected.GET("/threads/:id", threadHandler.GetThread)
			protected.PATCH("/threads/:id", threadHandler.UpdateThread)
			protected.DELETE("/threads/:id", threadHandler.DeleteThread)
			protected.POST("/threads/:id/archive", threadHandler.ArchiveThread)
			protected.POST("/threads/:id/unarchive", threadHandler.UnarchiveThread)
			// Voice message - with credit enforcement (1 credit per voice submission)
			protected.POST("/threads/:id/messages/audio",
				middleware.RequireCredits(creditsService, models.CreditCostPerMessage),
				threadHandler.SendAudioMessage)

			// Audio - use *key to capture full path including slashes
			protected.GET("/audio/*key", audioHandler.GetAudio)

			// Subscription and Credits
			protected.GET("/subscription", subscriptionHandler.GetSubscriptionStatus)
			protected.POST("/subscription/checkout", subscriptionHandler.CreateCheckoutSession)
			protected.POST("/subscription/portal", subscriptionHandler.CreatePortalSession)
			protected.GET("/credits", subscriptionHandler.GetCreditsBalance)
			protected.GET("/credits/history", subscriptionHandler.GetCreditHistory)

			// Pronunciation stats
			protected.GET("/pronunciation/stats", phonemeStatsHandler.GetStats)
		}

		// Stripe webhook (no auth - verified by Stripe signature)
		api.POST("/webhooks/stripe", subscriptionHandler.HandleStripeWebhook)
	}

	// Start server
	addr := cfg.Host + ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
		os.Exit(1)
	}
}
