package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	clientmocks "ling-app/api/internal/client/mocks"
	"ling-app/api/internal/middleware"
	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
	repomocks "ling-app/api/internal/repository/mocks"
	"ling-app/api/internal/services"
	servicemocks "ling-app/api/internal/services/mocks"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestThreadHandler_SendAudioMessage_Success(t *testing.T) {
	// Setup
	userID := uuid.New()
	threadID := uuid.New()

	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}

	thread := &models.Thread{
		ID:     threadID,
		UserID: userID,
	}

	userMessage := &models.Message{
		ID:       uuid.New(),
		ThreadID: threadID,
		Role:     "user",
		Content:  "hello",
		HasAudio: true,
	}

	assistantMessage := &models.Message{
		ID:       uuid.New(),
		ThreadID: threadID,
		Role:     "assistant",
		Content:  "Hi there!",
		HasAudio: true,
	}

	turn := &services.ConversationTurn{
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
	}

	// Mock repositories
	threadRepo := new(repomocks.MockThreadRepository)
	threadRepo.On("FindByIDAndUserID", mock.Anything, threadID, userID).Return(thread, nil)
	// For generateThreadName goroutine
	threadRepo.On("FindByID", mock.Anything, threadID).Return(thread, nil).Maybe()
	threadRepo.On("UpdateName", mock.Anything, threadID, mock.Anything).Return(nil).Maybe()

	// Mock OpenAI client for title generation
	openAIClient := new(clientmocks.MockOpenAIClient)
	openAIClient.On("GenerateTitle", mock.Anything).Return("Test Thread", nil).Maybe()

	// Mock conversation service
	conversationService := new(servicemocks.MockConversationProcessor)
	conversationService.On("ProcessAudioMessage", mock.Anything, threadID, mock.Anything, mock.Anything).
		Return(turn, nil)

	// Create handler
	handler := NewThreadHandler(nil, threadRepo, nil, conversationService, openAIClient, nil)

	// Setup router
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.UserContextKey, user)
		c.Next()
	})
	router.POST("/threads/:id/messages/audio", handler.SendAudioMessage)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("audio", "test.webm")
	assert.NoError(t, err)
	_, err = part.Write([]byte("fake audio data"))
	assert.NoError(t, err)
	writer.Close()

	// Execute request
	req := httptest.NewRequest("POST", "/threads/"+threadID.String()+"/messages/audio", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["userMessage"])
	assert.NotNil(t, response["assistantMessage"])

	// Verify mocks
	threadRepo.AssertExpectations(t)
	conversationService.AssertExpectations(t)
}

func TestThreadHandler_SendAudioMessage_InvalidThreadID(t *testing.T) {
	// Setup
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	handler := NewThreadHandler(nil, nil, nil, nil, nil, nil)

	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.UserContextKey, user)
		c.Next()
	})
	router.POST("/threads/:id/messages/audio", handler.SendAudioMessage)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("audio", "test.webm")
	assert.NoError(t, err)
	_, err = part.Write([]byte("fake audio data"))
	assert.NoError(t, err)
	writer.Close()

	// Execute with invalid UUID
	req := httptest.NewRequest("POST", "/threads/invalid-uuid/messages/audio", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid thread ID", response["error"])
}

func TestThreadHandler_SendAudioMessage_ThreadNotFound(t *testing.T) {
	// Setup
	userID := uuid.New()
	threadID := uuid.New()

	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}

	// Mock thread repo to return not found
	threadRepo := new(repomocks.MockThreadRepository)
	threadRepo.On("FindByIDAndUserID", mock.Anything, threadID, userID).
		Return(nil, repository.ErrNotFound)

	handler := NewThreadHandler(nil, threadRepo, nil, nil, nil, nil)

	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.UserContextKey, user)
		c.Next()
	})
	router.POST("/threads/:id/messages/audio", handler.SendAudioMessage)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("audio", "test.webm")
	assert.NoError(t, err)
	_, err = part.Write([]byte("fake audio data"))
	assert.NoError(t, err)
	writer.Close()

	// Execute
	req := httptest.NewRequest("POST", "/threads/"+threadID.String()+"/messages/audio", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Thread not found", response["error"])

	threadRepo.AssertExpectations(t)
}

func TestThreadHandler_SendAudioMessage_ThreadRepoError(t *testing.T) {
	// Setup
	userID := uuid.New()
	threadID := uuid.New()

	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}

	// Mock thread repo to return generic error
	threadRepo := new(repomocks.MockThreadRepository)
	threadRepo.On("FindByIDAndUserID", mock.Anything, threadID, userID).
		Return(nil, errors.New("database error"))

	handler := NewThreadHandler(nil, threadRepo, nil, nil, nil, nil)

	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.UserContextKey, user)
		c.Next()
	})
	router.POST("/threads/:id/messages/audio", handler.SendAudioMessage)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("audio", "test.webm")
	assert.NoError(t, err)
	_, err = part.Write([]byte("fake audio data"))
	assert.NoError(t, err)
	writer.Close()

	// Execute
	req := httptest.NewRequest("POST", "/threads/"+threadID.String()+"/messages/audio", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to fetch thread", response["error"])

	threadRepo.AssertExpectations(t)
}

func TestThreadHandler_SendAudioMessage_MissingAudioFile(t *testing.T) {
	// Setup
	userID := uuid.New()
	threadID := uuid.New()

	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}

	thread := &models.Thread{
		ID:     threadID,
		UserID: userID,
	}

	threadRepo := new(repomocks.MockThreadRepository)
	threadRepo.On("FindByIDAndUserID", mock.Anything, threadID, userID).Return(thread, nil)

	handler := NewThreadHandler(nil, threadRepo, nil, nil, nil, nil)

	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.UserContextKey, user)
		c.Next()
	})
	router.POST("/threads/:id/messages/audio", handler.SendAudioMessage)

	// Create empty request (no audio file)
	req := httptest.NewRequest("POST", "/threads/"+threadID.String()+"/messages/audio", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Audio file is required", response["error"])

	threadRepo.AssertExpectations(t)
}

func TestThreadHandler_SendAudioMessage_ProcessingError(t *testing.T) {
	// Setup
	userID := uuid.New()
	threadID := uuid.New()

	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}

	thread := &models.Thread{
		ID:     threadID,
		UserID: userID,
	}

	// Mock repositories
	threadRepo := new(repomocks.MockThreadRepository)
	threadRepo.On("FindByIDAndUserID", mock.Anything, threadID, userID).Return(thread, nil)

	// Mock conversation service to return error
	conversationService := new(servicemocks.MockConversationProcessor)
	conversationService.On("ProcessAudioMessage", mock.Anything, threadID, mock.Anything, mock.Anything).
		Return(nil, errors.New("processing failed"))

	handler := NewThreadHandler(nil, threadRepo, nil, conversationService, nil, nil)

	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.UserContextKey, user)
		c.Next()
	})
	router.POST("/threads/:id/messages/audio", handler.SendAudioMessage)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("audio", "test.webm")
	assert.NoError(t, err)
	_, err = part.Write([]byte("fake audio data"))
	assert.NoError(t, err)
	writer.Close()

	// Execute
	req := httptest.NewRequest("POST", "/threads/"+threadID.String()+"/messages/audio", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to process audio message", response["error"])

	threadRepo.AssertExpectations(t)
	conversationService.AssertExpectations(t)
}

// mockReadCloser wraps an io.Reader to add a Close method
type mockReadCloser struct {
	io.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}

func TestThreadHandler_GetThreads_Success(t *testing.T) {
	// Setup
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}

	threads := []models.Thread{
		{ID: uuid.New(), UserID: userID},
		{ID: uuid.New(), UserID: userID},
	}

	threadRepo := new(repomocks.MockThreadRepository)
	threadRepo.On("FindByUserID", mock.Anything, userID).Return(threads, nil)

	handler := NewThreadHandler(nil, threadRepo, nil, nil, nil, nil)

	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.UserContextKey, user)
		c.Next()
	})
	router.GET("/threads", handler.GetThreads)

	// Execute
	req := httptest.NewRequest("GET", "/threads", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Thread
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)

	threadRepo.AssertExpectations(t)
}

func TestThreadHandler_GetThreads_Error(t *testing.T) {
	// Setup
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}

	threadRepo := new(repomocks.MockThreadRepository)
	threadRepo.On("FindByUserID", mock.Anything, userID).Return(nil, errors.New("database error"))

	handler := NewThreadHandler(nil, threadRepo, nil, nil, nil, nil)

	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.UserContextKey, user)
		c.Next()
	})
	router.GET("/threads", handler.GetThreads)

	// Execute
	req := httptest.NewRequest("GET", "/threads", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	threadRepo.AssertExpectations(t)
}

func TestThreadHandler_GetArchivedThreads_Success(t *testing.T) {
	// Setup
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}

	archivedTime := time.Now()
	threads := []models.Thread{
		{ID: uuid.New(), UserID: userID, ArchivedAt: &archivedTime},
	}

	threadRepo := new(repomocks.MockThreadRepository)
	threadRepo.On("FindArchivedByUserID", mock.Anything, userID).Return(threads, nil)

	handler := NewThreadHandler(nil, threadRepo, nil, nil, nil, nil)

	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.UserContextKey, user)
		c.Next()
	})
	router.GET("/threads/archived", handler.GetArchivedThreads)

	// Execute
	req := httptest.NewRequest("GET", "/threads/archived", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Thread
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.NotNil(t, response[0].ArchivedAt)

	threadRepo.AssertExpectations(t)
}
