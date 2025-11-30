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
	MLServiceURL string

	// OpenAI
	OpenAIAPIKey string

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

		MLServiceURL: getEnv("ML_SERVICE_URL", "http://localhost:8000"),

		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),

		CORSAllowedOrigins: strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ","),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
