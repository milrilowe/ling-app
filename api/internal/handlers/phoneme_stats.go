package handlers

import (
	"net/http"

	"ling-app/api/internal/middleware"
	"ling-app/api/internal/services"

	"github.com/gin-gonic/gin"
)

type PhonemeStatsHandler struct {
	PhonemeStatsService services.PhonemeStatsProvider
}

func NewPhonemeStatsHandler(phonemeStatsService services.PhonemeStatsProvider) *PhonemeStatsHandler {
	return &PhonemeStatsHandler{
		PhonemeStatsService: phonemeStatsService,
	}
}

// GetStats returns aggregated phoneme statistics for the current user
// GET /api/pronunciation/stats
func (h *PhonemeStatsHandler) GetStats(c *gin.Context) {
	user := middleware.MustGetUser(c)

	stats, err := h.PhonemeStatsService.GetUserStats(user.ID)
	if err != nil {
		handleError(c, err, "GetStats")
		return
	}

	c.JSON(http.StatusOK, stats)
}
