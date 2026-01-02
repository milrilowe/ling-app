package main

import (
	"log"
	"os"
	"time"

	"ling-app/api/internal/client"
	"ling-app/api/internal/config"
	"ling-app/api/internal/db"
	"ling-app/api/internal/handlers"
	"ling-app/api/internal/middleware"
	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
	"ling-app/api/internal/services"
	"ling-app/api/internal/services/auth"

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

	// Initialize repositories
	userRepo := repository.NewUserRepository()
	sessionRepo := repository.NewSessionRepository()
	creditsRepo := repository.NewCreditsRepository()
	creditTxRepo := repository.NewCreditTransactionRepository()
	threadRepo := repository.NewThreadRepository()
	messageRepo := repository.NewMessageRepository()

	// Initialize auth services
	authService := auth.NewAuthService(database, userRepo, sessionRepo, cfg.SessionMaxAge)
	oauthService := services.NewOAuthService(cfg)

	// Initialize storage client
	storageClient, err := client.NewStorageClient(cfg.S3Bucket, cfg.S3Region)
	if err != nil {
		log.Fatal("Failed to initialize storage client:", err)
		os.Exit(1)
	}
	log.Printf("Storage bucket '%s' configured in region '%s'", cfg.S3Bucket, cfg.S3Region)

	// Initialize AI clients
	openAIClient := client.NewOpenAIClient(cfg.OpenAIAPIKey)
	whisperClient := client.NewWhisperClient(cfg.MLServiceURL) // Uses local faster-whisper via ML service
	ttsClient := client.NewTTSClient(cfg.MLServiceURL)         // Uses local Chatterbox TTS via ML service

	// Initialize ML client for pronunciation analysis
	mlClient := client.NewMLClient(cfg.MLServiceURL, time.Duration(cfg.MLServiceTimeout)*time.Second)

	// Initialize phoneme stats repositories and service
	phonemeStatsRepo := repository.NewPhonemeStatsRepository()
	phonemeSubsRepo := repository.NewPhonemeSubstitutionRepository()
	phonemeStatsService := services.NewPhonemeStatsService(database, phonemeStatsRepo, phonemeSubsRepo)

	// Initialize pronunciation worker
	pronunciationWorker := services.NewPronunciationWorker(database, mlClient, storageClient, phonemeStatsService)

	// Initialize credits and subscription services
	creditsService := services.NewCreditsService(database, creditsRepo, creditTxRepo)
	subscriptionRepo := repository.NewSubscriptionRepository()
	stripeService := services.NewStripeService(cfg, database, subscriptionRepo, creditsService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, oauthService, creditsService, cfg)
	threadHandler := handlers.NewThreadHandler(database.DB, threadRepo, messageRepo, openAIClient, storageClient, whisperClient, ttsClient, pronunciationWorker, creditsService, cfg.MaxAudioFileSize)
	audioHandler := handlers.NewAudioHandler(storageClient)
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
