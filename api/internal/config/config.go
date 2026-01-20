package config

import (
	"fmt"
	"log"
	"net/url"
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

	// TTS Service (empty = use OpenAI TTS, set to ML service URL for Chatterbox)
	TTSServiceURL string

	// STT Service (empty = use OpenAI Whisper, set to ML service URL for faster-whisper)
	STTServiceURL string

	// OpenAI
	OpenAIAPIKey string

	// S3/MinIO Storage
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3Region    string

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

		TTSServiceURL: getEnv("TTS_SERVICE_URL", ""), // Empty = OpenAI TTS, or set to ML service URL

		STTServiceURL: getEnv("STT_SERVICE_URL", ""), // Empty = OpenAI Whisper, or set to ML service URL

		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),

		S3Endpoint:  getEnv("S3_ENDPOINT", "http://localhost:9000"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey: getEnv("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:    getEnv("S3_BUCKET", "ling-app-audio"),
		S3Region:    getEnv("S3_REGION", "us-east-1"),

		MaxAudioFileSize: 10485760, // 10MB

		CORSAllowedOrigins: strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000"), ","),

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

// Validate checks that required configuration values are set and valid
func (c *Config) Validate() error {
	// Required fields
	required := map[string]string{
		"DATABASE_URL":    c.DatabaseURL,
		"SESSION_SECRET":  c.SessionSecret,
		"S3_ENDPOINT":     c.S3Endpoint,
		"S3_ACCESS_KEY":   c.S3AccessKey,
		"S3_SECRET_KEY":   c.S3SecretKey,
		"S3_BUCKET":       c.S3Bucket,
	}

	for name, value := range required {
		if value == "" {
			return fmt.Errorf("required config %s is not set", name)
		}
	}

	// Validate DATABASE_URL format
	if _, err := url.Parse(c.DatabaseURL); err != nil {
		return fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	// Validate SESSION_SECRET length
	if len(c.SessionSecret) < 32 {
		return fmt.Errorf("SESSION_SECRET must be at least 32 characters for security")
	}

	// Validate that at least one of STT/TTS service URL or OpenAI key is set
	if c.STTServiceURL == "" && c.OpenAIAPIKey == "" {
		return fmt.Errorf("either STT_SERVICE_URL or OPENAI_API_KEY must be set for speech-to-text")
	}

	if c.TTSServiceURL == "" && c.OpenAIAPIKey == "" {
		return fmt.Errorf("either TTS_SERVICE_URL or OPENAI_API_KEY must be set for text-to-speech")
	}

	return nil
}
