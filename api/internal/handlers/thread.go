package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"ling-app/api/internal/db"
	"ling-app/api/internal/middleware"
	"ling-app/api/internal/models"
	"ling-app/api/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ThreadHandler struct {
	DB                  *db.DB
	OpenAIClient        *services.OpenAIClient
	Storage             *services.StorageService
	WhisperClient       *services.WhisperClient
	ElevenLabsClient    *services.ElevenLabsClient
	PronunciationWorker *services.PronunciationWorker
	CreditsService      *services.CreditsService
	MaxAudioFileSize    int64
}

func NewThreadHandler(database *db.DB, openAIClient *services.OpenAIClient, storage *services.StorageService, whisper *services.WhisperClient, elevenlabs *services.ElevenLabsClient, pronunciationWorker *services.PronunciationWorker, creditsService *services.CreditsService, maxAudioFileSize int64) *ThreadHandler {
	return &ThreadHandler{
		DB:                  database,
		OpenAIClient:        openAIClient,
		Storage:             storage,
		WhisperClient:       whisper,
		ElevenLabsClient:    elevenlabs,
		PronunciationWorker: pronunciationWorker,
		CreditsService:      creditsService,
		MaxAudioFileSize:    maxAudioFileSize,
	}
}

type CreateThreadRequest struct {
	InitialPrompt    string `json:"initialPrompt"`
	FirstUserMessage string `json:"firstUserMessage"`
}


// GetThreads retrieves all threads for the current user, ordered by most recent
func (h *ThreadHandler) GetThreads(c *gin.Context) {
	// Get current user from context (set by RequireAuth middleware)
	user := middleware.MustGetUser(c)

	var threads []models.Thread
	if err := h.DB.Where("user_id = ?", user.ID).Order("created_at DESC").Find(&threads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch threads"})
		return
	}
	c.JSON(http.StatusOK, threads)
}

// CreateThread creates a new conversation thread
func (h *ThreadHandler) CreateThread(c *gin.Context) {
	// Get current user from context
	user := middleware.MustGetUser(c)

	var req CreateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	thread := models.Thread{
		ID:        uuid.New(),
		UserID:    user.ID, // Associate thread with user
		CreatedAt: time.Now(),
	}

	// Create thread
	if err := h.DB.Create(&thread).Error; err != nil {
		log.Printf("Error creating thread: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create thread"})
		return
	}

	// Add initial AI prompt message only if provided
	if req.InitialPrompt != "" {
		aiMessage := models.Message{
			ID:        uuid.New(),
			ThreadID:  thread.ID,
			Role:      "assistant",
			Content:   req.InitialPrompt,
			Timestamp: time.Now(),
		}

		if err := h.DB.Create(&aiMessage).Error; err != nil {
			log.Printf("Error creating AI message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
			return
		}
	}

	// Add first user message if provided
	if req.FirstUserMessage != "" {
		userMessage := models.Message{
			ID:        uuid.New(),
			ThreadID:  thread.ID,
			Role:      "user",
			Content:   req.FirstUserMessage,
			Timestamp: time.Now(),
		}

		if err := h.DB.Create(&userMessage).Error; err != nil {
			log.Printf("Error creating user message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
			return
		}

		// Generate AI response
		conversationHistory := []services.ConversationMessage{}
		if req.InitialPrompt != "" {
			conversationHistory = append(conversationHistory, services.ConversationMessage{
				Role:    "assistant",
				Content: req.InitialPrompt,
			})
		}
		conversationHistory = append(conversationHistory, services.ConversationMessage{
			Role:    "user",
			Content: req.FirstUserMessage,
		})

		aiResponse, err := h.OpenAIClient.Generate(conversationHistory)
		if err != nil {
			log.Printf("Error generating AI response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate response"})
			return
		}

		responseMessage := models.Message{
			ID:        uuid.New(),
			ThreadID:  thread.ID,
			Role:      "assistant",
			Content:   aiResponse,
			Timestamp: time.Now(),
		}

		if err := h.DB.Create(&responseMessage).Error; err != nil {
			log.Printf("Error creating AI response message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
			return
		}
	}

	// Load thread with messages
	if err := h.DB.Preload("Messages").First(&thread, thread.ID).Error; err != nil {
		log.Printf("Error loading thread: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load thread"})
		return
	}

	c.JSON(http.StatusOK, thread)
}

// GetThread retrieves a thread with all messages (only if owned by current user)
func (h *ThreadHandler) GetThread(c *gin.Context) {
	user := middleware.MustGetUser(c)
	threadID := c.Param("id")

	parsedID, err := uuid.Parse(threadID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	var thread models.Thread
	// Only fetch if thread belongs to current user
	if err := h.DB.Preload("Messages").Where("id = ? AND user_id = ?", parsedID, user.ID).First(&thread).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	c.JSON(http.StatusOK, thread)
}

// SendAudioMessage handles audio message upload, transcription, AI response, and TTS
// POST /api/threads/:id/messages/audio
func (h *ThreadHandler) SendAudioMessage(c *gin.Context) {
	user := middleware.MustGetUser(c)
	threadID := c.Param("id")
	parsedID, err := uuid.Parse(threadID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	// Check if thread exists and belongs to current user
	var thread models.Thread
	if err := h.DB.Where("id = ? AND user_id = ?", parsedID, user.ID).First(&thread).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	// Get audio file from multipart form
	file, fileHeader, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio file is required"})
		return
	}
	defer file.Close()

	// Validate file size
	if h.MaxAudioFileSize > 0 && fileHeader.Size > h.MaxAudioFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error":   "Audio file too large",
			"maxSize": h.MaxAudioFileSize,
		})
		return
	}

	// Create user message ID
	userMessageID := uuid.New()

	// Upload user audio to storage
	userAudioKey := fmt.Sprintf("user/%s/%s.webm", parsedID, userMessageID)
	ctx := context.Background()
	_, err = h.Storage.UploadAudio(ctx, file, userAudioKey, "audio/webm")
	if err != nil {
		log.Printf("Error uploading user audio: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload audio"})
		return
	}

	// Get presigned URL for ML service to access the audio
	audioPresignedURL, err := h.Storage.GetPresignedURL(ctx, userAudioKey, 5*time.Minute)
	if err != nil {
		log.Printf("Error getting presigned URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process audio"})
		return
	}

	// Transcribe audio using local Whisper via ML service
	transcription, err := h.WhisperClient.TranscribeFromURL(ctx, audioPresignedURL)
	if err != nil {
		log.Printf("Error transcribing audio: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transcribe audio"})
		return
	}

	// Save user message with audio (pronunciation analysis pending)
	userMessage := models.Message{
		ID:                    userMessageID,
		ThreadID:              parsedID,
		Role:                  "user",
		Content:               transcription.Text,
		AudioURL:              &userAudioKey,
		AudioDurationSeconds:  &transcription.Duration,
		HasAudio:              true,
		Timestamp:             time.Now(),
		PronunciationStatus:   "pending",
	}

	if err := h.DB.Create(&userMessage).Error; err != nil {
		log.Printf("Error creating user message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	// Spawn pronunciation analysis in background (non-blocking)
	if h.PronunciationWorker != nil {
		go h.PronunciationWorker.AnalyzeAsync(userMessageID, userAudioKey, transcription.Text, "en-us")
	}

	// Get conversation history
	var messages []models.Message
	if err := h.DB.Where("thread_id = ?", parsedID).Order("timestamp ASC").Find(&messages).Error; err != nil {
		log.Printf("Error fetching messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	// Convert to OpenAI format
	conversationHistory := make([]services.ConversationMessage, len(messages))
	for i, msg := range messages {
		conversationHistory[i] = services.ConversationMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Generate AI response
	aiResponse, err := h.OpenAIClient.Generate(conversationHistory)
	if err != nil {
		log.Printf("Error generating AI response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate response"})
		return
	}

	// Generate TTS for AI response
	assistantMessageID := uuid.New()
	ttsResult, err := h.ElevenLabsClient.TextToSpeech(aiResponse, "")
	if err != nil {
		log.Printf("Error generating TTS: %v", err)
		// Continue without audio - save text-only response
		responseMessage := models.Message{
			ID:        assistantMessageID,
			ThreadID:  parsedID,
			Role:      "assistant",
			Content:   aiResponse,
			HasAudio:  false,
			Timestamp: time.Now(),
		}

		if err := h.DB.Create(&responseMessage).Error; err != nil {
			log.Printf("Error creating AI response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"userMessage":      userMessage,
			"assistantMessage": responseMessage,
		})
		return
	}

	// Upload TTS audio to storage
	assistantAudioKey := fmt.Sprintf("assistant/%s/%s.mp3", parsedID, assistantMessageID)
	audioReader := bytes.NewReader(ttsResult.AudioBytes)
	_, err = h.Storage.UploadAudio(ctx, audioReader, assistantAudioKey, "audio/mpeg")
	if err != nil {
		log.Printf("Error uploading TTS audio: %v", err)
		// Continue without audio
		responseMessage := models.Message{
			ID:        assistantMessageID,
			ThreadID:  parsedID,
			Role:      "assistant",
			Content:   aiResponse,
			HasAudio:  false,
			Timestamp: time.Now(),
		}

		if err := h.DB.Create(&responseMessage).Error; err != nil {
			log.Printf("Error creating AI response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"userMessage":      userMessage,
			"assistantMessage": responseMessage,
		})
		return
	}

	// Save AI response with audio
	ttsDuration := ttsResult.Duration
	responseMessage := models.Message{
		ID:                   assistantMessageID,
		ThreadID:             parsedID,
		Role:                 "assistant",
		Content:              aiResponse,
		AudioURL:             &assistantAudioKey,
		AudioDurationSeconds: &ttsDuration,
		HasAudio:             true,
		Timestamp:            time.Now(),
	}

	if err := h.DB.Create(&responseMessage).Error; err != nil {
		log.Printf("Error creating AI response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response"})
		return
	}

	// Deduct credits for voice message
	if h.CreditsService != nil {
		cost := middleware.GetCreditsCost(c)
		if cost > 0 {
			if err := h.CreditsService.DeductCredits(user.ID, cost, responseMessage.ID.String(), "Voice message"); err != nil {
				// Credit deduction failed - this is a billing issue that needs attention
				// The message was already processed, so we return it but log the error prominently
				log.Printf("CRITICAL: Failed to deduct credits for user %s, message %s: %v", user.ID, responseMessage.ID, err)
				// Still return success since the message was processed - but this needs monitoring
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"userMessage":      userMessage,
		"assistantMessage": responseMessage,
	})
}
