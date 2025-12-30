package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	clientmocks "ling-app/api/internal/client/mocks"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAudioHandler_GetAudio(t *testing.T) {
	t.Run("returns presigned URL successfully", func(t *testing.T) {
		storageClient := new(clientmocks.MockStorageClient)
		handler := NewAudioHandler(storageClient)

		storageClient.On("GetPresignedURL", mock.Anything, "audio/test.wav", 24*time.Hour).
			Return("https://presigned.url/audio/test.wav", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "key", Value: "/audio/test.wav"}}

		handler.GetAudio(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "https://presigned.url/audio/test.wav", response["url"])
		storageClient.AssertExpectations(t)
	})

	t.Run("strips leading slash from key", func(t *testing.T) {
		storageClient := new(clientmocks.MockStorageClient)
		handler := NewAudioHandler(storageClient)

		// The key passed to storage should have the leading slash removed
		storageClient.On("GetPresignedURL", mock.Anything, "user/123/audio.wav", 24*time.Hour).
			Return("https://presigned.url/user/123/audio.wav", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "key", Value: "/user/123/audio.wav"}}

		handler.GetAudio(c)

		assert.Equal(t, http.StatusOK, w.Code)
		storageClient.AssertExpectations(t)
	})

	t.Run("returns error when key is empty", func(t *testing.T) {
		storageClient := new(clientmocks.MockStorageClient)
		handler := NewAudioHandler(storageClient)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "key", Value: ""}}

		handler.GetAudio(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Audio key is required", response["error"])
		storageClient.AssertNotCalled(t, "GetPresignedURL")
	})

	t.Run("returns error when storage fails", func(t *testing.T) {
		storageClient := new(clientmocks.MockStorageClient)
		handler := NewAudioHandler(storageClient)

		storageClient.On("GetPresignedURL", mock.Anything, "audio/test.wav", 24*time.Hour).
			Return("", errors.New("storage unavailable"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "key", Value: "/audio/test.wav"}}

		handler.GetAudio(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Failed to generate audio URL")
	})
}
