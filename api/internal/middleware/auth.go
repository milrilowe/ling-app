package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ling-app/api/internal/models"
	"ling-app/api/internal/services/auth"
)

// Context keys for storing user data
const (
	UserContextKey = "user"
)

// RequireAuth is middleware that requires a valid session.
// If the session is invalid or missing, it returns 401 Unauthorized.
// If valid, it sets the user in the Gin context for handlers to access.
func RequireAuth(authService *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session token from cookie
		token, err := c.Cookie("session_token")
		if err != nil {
			// No cookie = not authenticated
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			return
		}

		// Validate session and get user
		user, err := authService.ValidateSession(token)
		if err != nil {
			// Invalid or expired session
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired session",
			})
			return
		}

		// Store user in context for handlers to access
		c.Set(UserContextKey, user)

		// Continue to the next handler
		c.Next()
	}
}

// OptionalAuth is middleware that checks for authentication but doesn't require it.
// If a valid session exists, the user is set in context.
// If not, the request continues without a user.
// Useful for endpoints that work differently for authenticated vs anonymous users.
func OptionalAuth(authService *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil {
			// No cookie, continue without user
			c.Next()
			return
		}

		user, err := authService.ValidateSession(token)
		if err != nil {
			// Invalid session, continue without user
			c.Next()
			return
		}

		// Valid session, set user in context
		c.Set(UserContextKey, user)
		c.Next()
	}
}

// GetUserFromContext retrieves the authenticated user from the Gin context.
// Returns the user and true if authenticated, or nil and false if not.
// Use this in handlers after RequireAuth middleware.
func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	value, exists := c.Get(UserContextKey)
	if !exists {
		return nil, false
	}

	user, ok := value.(*models.User)
	if !ok {
		return nil, false
	}

	return user, true
}

// MustGetUser is like GetUserFromContext but panics if user is not found.
// Only use this after RequireAuth middleware where you're certain the user exists.
func MustGetUser(c *gin.Context) *models.User {
	user, ok := GetUserFromContext(c)
	if !ok {
		panic("MustGetUser called without authenticated user in context")
	}
	return user
}
