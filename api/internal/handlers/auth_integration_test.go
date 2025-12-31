//go:build integration

package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ling-app/api/internal/config"
	"ling-app/api/internal/handlers"
	"ling-app/api/internal/repository"
	"ling-app/api/internal/services"
	"ling-app/api/internal/services/auth"
	"ling-app/api/internal/testutil"
)

func setupAuthTestRouter(t *testing.T) (*gin.Engine, *testutil.TestDB) {
	gin.SetMode(gin.TestMode)

	testDB := testutil.NewTestDB(t)
	if testDB == nil {
		return nil, nil
	}

	t.Cleanup(func() {
		testDB.Cleanup()
	})

	// Initialize repositories
	userRepo := repository.NewUserRepository()
	sessionRepo := repository.NewSessionRepository()
	creditsRepo := repository.NewCreditsRepository()
	creditTxRepo := repository.NewCreditTransactionRepository()

	// Initialize services
	authService := auth.NewAuthService(testDB.DB, userRepo, sessionRepo, 86400)
	creditsService := services.NewCreditsService(testDB.DB, creditsRepo, creditTxRepo)

	// Config for testing
	cfg := &config.Config{
		SessionMaxAge: 86400,
		Environment:   "test",
		FrontendURL:   "http://localhost:3000",
	}

	// Initialize handler
	authHandler := handlers.NewAuthHandler(authService, nil, creditsService, cfg)

	// Setup router
	router := gin.New()
	api := router.Group("/api/auth")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.POST("/logout", authHandler.Logout)
	}

	return router, testDB
}

func TestRegisterIntegration(t *testing.T) {
	router, testDB := setupAuthTestRouter(t)
	if testDB == nil {
		return
	}

	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "successful registration",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
				"name":     "Test User",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "test@example.com", body["email"])
				assert.Equal(t, "Test User", body["name"])
				assert.NotEmpty(t, body["id"])
			},
		},
		{
			name: "duplicate email",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "password456",
				"name":     "Another User",
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "already registered")
			},
		},
		{
			name: "invalid email",
			payload: map[string]string{
				"email":    "not-an-email",
				"password": "password123",
				"name":     "Test User",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "password too short",
			payload: map[string]string{
				"email":    "short@example.com",
				"password": "short",
				"name":     "Test User",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing name",
			payload: map[string]string{
				"email":    "noname@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				var respBody map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &respBody)
				require.NoError(t, err)
				tt.checkResponse(t, respBody)
			}
		})
	}
}

func TestLoginIntegration(t *testing.T) {
	router, testDB := setupAuthTestRouter(t)
	if testDB == nil {
		return
	}

	// First, register a user
	registerPayload, _ := json.Marshal(map[string]string{
		"email":    "login@example.com",
		"password": "password123",
		"name":     "Login Test User",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(registerPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
		checkCookie    bool
	}{
		{
			name: "successful login",
			payload: map[string]string{
				"email":    "login@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
			checkCookie:    true,
		},
		{
			name: "wrong password",
			payload: map[string]string{
				"email":    "login@example.com",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "non-existent user",
			payload: map[string]string{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "case insensitive email",
			payload: map[string]string{
				"email":    "LOGIN@EXAMPLE.COM",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
			checkCookie:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkCookie {
				cookies := w.Result().Cookies()
				var sessionCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "session_token" {
						sessionCookie = c
						break
					}
				}
				assert.NotNil(t, sessionCookie, "session cookie should be set")
				assert.NotEmpty(t, sessionCookie.Value)
			}
		})
	}
}

func TestLogoutIntegration(t *testing.T) {
	router, testDB := setupAuthTestRouter(t)
	if testDB == nil {
		return
	}

	// Register and login first
	registerPayload, _ := json.Marshal(map[string]string{
		"email":    "logout@example.com",
		"password": "password123",
		"name":     "Logout Test User",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(registerPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Get the session cookie
	var sessionCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "session_token" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie)

	// Now logout
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutW := httptest.NewRecorder()
	router.ServeHTTP(logoutW, logoutReq)

	assert.Equal(t, http.StatusOK, logoutW.Code)

	// Check that cookie is cleared
	for _, c := range logoutW.Result().Cookies() {
		if c.Name == "session_token" {
			assert.Equal(t, "", c.Value, "session cookie should be cleared")
			assert.True(t, c.MaxAge < 0, "cookie should be expired")
		}
	}
}

func TestCreditsInitializedOnRegister(t *testing.T) {
	router, testDB := setupAuthTestRouter(t)
	if testDB == nil {
		return
	}

	// Register a user
	payload, _ := json.Marshal(map[string]string{
		"email":    "credits@example.com",
		"password": "password123",
		"name":     "Credits Test User",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var respBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &respBody)
	require.NoError(t, err)

	userID := respBody["id"].(string)

	// Check credits were created in database
	var count int64
	testDB.Raw("SELECT COUNT(*) FROM credits WHERE user_id = ?", userID).Scan(&count)
	assert.Equal(t, int64(1), count, "credits record should be created for new user")
}
