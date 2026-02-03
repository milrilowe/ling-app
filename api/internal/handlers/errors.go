package handlers

import (
	"errors"
	"log"
	"net/http"

	"ling-app/api/internal/repository"
	"ling-app/api/internal/services"
	"ling-app/api/internal/services/auth"

	"github.com/gin-gonic/gin"
)

// handleError maps common errors to appropriate HTTP responses
func handleError(c *gin.Context, err error, operation string) {
	// Log the error with context
	log.Printf("[%s] Error: %v", operation, err)

	// Map errors to HTTP status codes
	switch {
	// Repository errors
	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})

	// Auth errors
	case errors.Is(err, auth.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	case errors.Is(err, auth.ErrEmailTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
	case errors.Is(err, auth.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	case errors.Is(err, auth.ErrSessionNotFound):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found"})

	// Service errors
	case errors.Is(err, services.ErrSubscriptionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
	case errors.Is(err, services.ErrInvalidWebhook):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook"})
	case errors.Is(err, services.ErrInsufficientCredits):
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient credits"})

	// Validation errors
	case errors.Is(err, services.ErrAudioTooShort):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio must be at least 1 second long. Please record a longer message."})
	case errors.Is(err, services.ErrAudioTooLong):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio must be 30 seconds or less. Please record a shorter message."})
	case errors.Is(err, services.ErrAudioInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audio file"})

	// Default to internal server error
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

// handleValidationError handles request validation/binding errors
func handleValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

// handleNotFound is a convenience function for 404 responses
func handleNotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, gin.H{"error": resource + " not found"})
}
