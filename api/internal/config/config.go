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

	// ML Service
	MLServiceURL     string
	MLServiceTimeout int // timeout in seconds for ML service calls

	// OpenAI
	OpenAIAPIKey string

	// ElevenLabs
	ElevenLabsAPIKey string

	// S3/MinIO Storage
	S3Endpoint   string
	S3AccessKey  string
	S3SecretKey  string
	S3Bucket     string
	S3Region     string

	// Audio Limits
	MaxAudioFileSize int64

	// CORS
	CORSAllowedOrigins []string
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

		MLServiceURL:     getEnv("ML_SERVICE_URL", "http://localhost:8000"),
		MLServiceTimeout: 120, // 2 minutes for pronunciation analysis

		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),

		ElevenLabsAPIKey: getEnv("ELEVENLABS_API_KEY", ""),

		S3Endpoint:  getEnv("S3_ENDPOINT", "http://localhost:9000"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey: getEnv("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:    getEnv("S3_BUCKET", "ling-app-audio"),
		S3Region:    getEnv("S3_REGION", "us-east-1"),

		MaxAudioFileSize: 10485760, // 10MB

		CORSAllowedOrigins: strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ","),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
