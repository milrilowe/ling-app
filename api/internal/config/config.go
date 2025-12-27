package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port     string
	Host     string
	GinMode  string
	Environment string

	// Database
	DatabaseURL string

	// Session
	SessionSecret string
	SessionMaxAge int

	// OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
	FrontendURL        string // Where to redirect after OAuth

	// ML Service
	MLServiceURL     string
	MLServiceTimeout int // timeout in seconds for ML service calls

	// MFA (Montreal Forced Aligner) Service
	MFAServiceURL     string
	MFAServiceTimeout int // timeout in seconds for alignment

	// OpenAI
	OpenAIAPIKey string

	// ElevenLabs
	ElevenLabsAPIKey string

	// S3/MinIO Storage
	S3Endpoint         string
	S3InternalEndpoint string // Internal endpoint for Docker-to-Docker communication (e.g., http://minio:9000)
	S3AccessKey        string
	S3SecretKey        string
	S3Bucket           string
	S3Region           string

	// Audio Limits
	MaxAudioFileSize int64

	// CORS
	CORSAllowedOrigins []string

	// Stripe
	StripeSecretKey     string
	StripeWebhookSecret string
	StripePriceBasic    string
	StripePricePro      string
	StripeSuccessURL    string
	StripeCancelURL     string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Port:     getEnv("PORT", "8080"),
		Host:     getEnv("HOST", "0.0.0.0"),
		GinMode:  getEnv("GIN_MODE", "debug"),
		Environment: getEnv("ENVIRONMENT", "development"),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		SessionSecret: getEnv("SESSION_SECRET", ""),
		SessionMaxAge: 86400, // 24 hours

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/api/auth/github/callback"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),

		MLServiceURL:     getEnv("ML_SERVICE_URL", "http://localhost:8000"),
		MLServiceTimeout: 120, // 2 minutes for pronunciation analysis

		MFAServiceURL:     getEnv("MFA_SERVICE_URL", "http://localhost:8001"),
		MFAServiceTimeout: 120, // 2 minutes for alignment (MFA can be slow)

		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),

		ElevenLabsAPIKey: getEnv("ELEVENLABS_API_KEY", ""),

		S3Endpoint:         getEnv("S3_ENDPOINT", "http://localhost:9000"),
		S3InternalEndpoint: getEnv("S3_INTERNAL_ENDPOINT", "http://minio:9000"), // For Docker-to-Docker communication
		S3AccessKey:        getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey:        getEnv("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:           getEnv("S3_BUCKET", "ling-app-audio"),
		S3Region:           getEnv("S3_REGION", "us-east-1"),

		MaxAudioFileSize: 10485760, // 10MB

		CORSAllowedOrigins: strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ","),

		StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripePriceBasic:    getEnv("STRIPE_PRICE_BASIC", ""),
		StripePricePro:      getEnv("STRIPE_PRICE_PRO", ""),
		StripeSuccessURL:    getEnv("STRIPE_SUCCESS_URL", "http://localhost:3000/subscription/success"),
		StripeCancelURL:     getEnv("STRIPE_CANCEL_URL", "http://localhost:3000/pricing"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
