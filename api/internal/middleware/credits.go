package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ling-app/api/internal/services"
)

// Context keys for credit cost
const (
	CreditsCostContextKey = "credits_cost"
)

// RequireCredits is middleware that checks if the user has enough credits.
// If they don't, it returns 402 Payment Required with INSUFFICIENT_CREDITS error code.
// The cost is stored in context for the handler to use for deduction after success.
func RequireCredits(creditsService *services.CreditsService, amount int) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := MustGetUser(c)

		hasCredits, err := creditsService.HasCredits(user.ID, amount)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check credits",
			})
			return
		}

		if !hasCredits {
			c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{
				"error":         "Insufficient credits",
				"code":          "INSUFFICIENT_CREDITS",
				"creditsNeeded": amount,
			})
			return
		}

		// Store the cost in context for the handler to deduct after success
		c.Set(CreditsCostContextKey, amount)
		c.Next()
	}
}

// GetCreditsCost retrieves the credit cost from context.
// Returns 0 if not set.
func GetCreditsCost(c *gin.Context) int {
	value, exists := c.Get(CreditsCostContextKey)
	if !exists {
		return 0
	}

	cost, ok := value.(int)
	if !ok {
		return 0
	}

	return cost
}
