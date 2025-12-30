package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"ling-app/api/internal/client"

	"github.com/gin-gonic/gin"
)

type AudioHandler struct {
	Storage client.StorageClient
}

func NewAudioHandler(storage client.StorageClient) *AudioHandler {
	return &AudioHandler{
		Storage: storage,
	}
}

// GetAudio generates a presigned URL for audio playback
// GET /api/audio/*key
func (h *AudioHandler) GetAudio(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio key is required"})
		return
	}

	// Remove leading slash from wildcard parameter
	if len(key) > 0 && key[0] == '/' {
		key = key[1:]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Generate presigned URL valid for 24 hours
	url, err := h.Storage.GetPresignedURL(ctx, key, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate audio URL: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}
