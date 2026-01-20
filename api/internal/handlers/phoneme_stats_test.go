package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"ling-app/api/internal/middleware"
	"ling-app/api/internal/models"
	"ling-app/api/internal/services"
	servicemocks "ling-app/api/internal/services/mocks"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPhonemeStatsHandler_GetStats_Success(t *testing.T) {
	// Setup
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}

	expectedStats := &services.UserPhonemeStatsResponse{
		TotalPhonemes:   100,
		OverallAccuracy: 90.0,
		PhonemeStats: []services.PhonemeAccuracy{
			{
				Phoneme:       "æ",
				TotalAttempts: 20,
				CorrectCount:  18,
				DeletionCount: 2,
				Accuracy:      90.0,
			},
		},
		CommonSubstitutions: []services.SubstitutionPattern{
			{
				ExpectedPhoneme: "θ",
				ActualPhoneme:   "s",
				Count:           5,
			},
		},
	}

	// Mock service
	phonemeService := new(servicemocks.MockPhonemeStatsProvider)
	phonemeService.On("GetUserStats", userID).Return(expectedStats, nil)

	handler := NewPhonemeStatsHandler(phonemeService)

	// Setup router
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.UserContextKey, user)
		c.Next()
	})
	router.GET("/stats", handler.GetStats)

	// Execute
	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response services.UserPhonemeStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedStats.OverallAccuracy, response.OverallAccuracy)
	assert.Equal(t, expectedStats.TotalPhonemes, response.TotalPhonemes)
	assert.Len(t, response.PhonemeStats, 1)
	assert.Len(t, response.CommonSubstitutions, 1)

	phonemeService.AssertExpectations(t)
}

func TestPhonemeStatsHandler_GetStats_ServiceError(t *testing.T) {
	// Setup
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}

	// Mock service to return error
	phonemeService := new(servicemocks.MockPhonemeStatsProvider)
	phonemeService.On("GetUserStats", userID).Return(nil, errors.New("database error"))

	handler := NewPhonemeStatsHandler(phonemeService)

	// Setup router
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.UserContextKey, user)
		c.Next()
	})
	router.GET("/stats", handler.GetStats)

	// Execute
	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	phonemeService.AssertExpectations(t)
}

func TestPhonemeStatsHandler_GetStats_EmptyStats(t *testing.T) {
	// Setup
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}

	// Mock service to return empty stats
	emptyStats := &services.UserPhonemeStatsResponse{
		TotalPhonemes:       0,
		OverallAccuracy:     0.0,
		PhonemeStats:        []services.PhonemeAccuracy{},
		CommonSubstitutions: []services.SubstitutionPattern{},
	}

	phonemeService := new(servicemocks.MockPhonemeStatsProvider)
	phonemeService.On("GetUserStats", userID).Return(emptyStats, nil)

	handler := NewPhonemeStatsHandler(phonemeService)

	// Setup router
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.UserContextKey, user)
		c.Next()
	})
	router.GET("/stats", handler.GetStats)

	// Execute
	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response services.UserPhonemeStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.TotalPhonemes)
	assert.Equal(t, 0.0, response.OverallAccuracy)
	assert.Empty(t, response.PhonemeStats)
	assert.Empty(t, response.CommonSubstitutions)

	phonemeService.AssertExpectations(t)
}
