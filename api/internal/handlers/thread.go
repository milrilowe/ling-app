package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"ling-app/api/internal/client"
	"ling-app/api/internal/middleware"
	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
	"ling-app/api/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ThreadHandler struct {
	exec                repository.Executor
	threadRepo          repository.ThreadRepository
	messageRepo         repository.MessageRepository
	conversationService services.ConversationProcessor
	OpenAIClient        client.OpenAIClient
	CreditsService      *services.CreditsService
}

func NewThreadHandler(
	exec repository.Executor,
	threadRepo repository.ThreadRepository,
	messageRepo repository.MessageRepository,
	conversationService services.ConversationProcessor,
	openAIClient client.OpenAIClient,
	creditsService *services.CreditsService,
) *ThreadHandler {
	return &ThreadHandler{
		exec:                exec,
		threadRepo:          threadRepo,
		messageRepo:         messageRepo,
		conversationService: conversationService,
		OpenAIClient:        openAIClient,
		CreditsService:      creditsService,
	}
}

type CreateThreadRequest struct {
	InitialPrompt    string `json:"initialPrompt"`
	FirstUserMessage string `json:"firstUserMessage"`
}

// GetThreads retrieves all non-archived threads for the current user, ordered by most recent
func (h *ThreadHandler) GetThreads(c *gin.Context) {
	user := middleware.MustGetUser(c)

	threads, err := h.threadRepo.FindByUserID(h.exec, user.ID)
	if err != nil {
		handleError(c, err, "GetThreads")
		return
	}
	c.JSON(http.StatusOK, threads)
}

// GetArchivedThreads retrieves all archived threads for the current user
func (h *ThreadHandler) GetArchivedThreads(c *gin.Context) {
	user := middleware.MustGetUser(c)

	threads, err := h.threadRepo.FindArchivedByUserID(h.exec, user.ID)
	if err != nil {
		handleError(c, err, "GetArchivedThreads")
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
	if err := h.threadRepo.Create(h.exec, &thread); err != nil {
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

		if err := h.messageRepo.Create(h.exec, &aiMessage); err != nil {
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

		if err := h.messageRepo.Create(h.exec, &userMessage); err != nil {
			log.Printf("Error creating user message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
			return
		}

		// Generate AI response
		conversationHistory := []client.ConversationMessage{}
		if req.InitialPrompt != "" {
			conversationHistory = append(conversationHistory, client.ConversationMessage{
				Role:    "assistant",
				Content: req.InitialPrompt,
			})
		}
		conversationHistory = append(conversationHistory, client.ConversationMessage{
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

		if err := h.messageRepo.Create(h.exec, &responseMessage); err != nil {
			log.Printf("Error creating AI response message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
			return
		}

		// Auto-generate thread name from AI response (async)
		go h.generateThreadName(thread.ID, aiResponse)
	}

	// Load thread with messages
	threadWithMessages, err := h.threadRepo.FindByIDWithMessages(h.exec, thread.ID)
	if err != nil {
		log.Printf("Error loading thread: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load thread"})
		return
	}

	c.JSON(http.StatusOK, threadWithMessages)
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

	thread, err := h.threadRepo.FindByIDAndUserIDWithMessages(h.exec, parsedID, user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch thread"})
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
	thread, err := h.threadRepo.FindByIDAndUserID(h.exec, parsedID, user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch thread"})
		return
	}

	// Get audio file from multipart form
	file, fileHeader, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio file is required"})
		return
	}
	defer file.Close()

	// Process audio message via ConversationService
	turn, err := h.conversationService.ProcessAudioMessage(c.Request.Context(), parsedID, file, fileHeader)
	if err != nil {
		handleError(c, err, "ProcessAudioMessage")
		return
	}

	// Auto-generate thread name from AI response (async)
	go h.generateThreadName(thread.ID, turn.AssistantMessage.Content)

	// Deduct credits for voice message
	if h.CreditsService != nil {
		cost := middleware.GetCreditsCost(c)
		if cost > 0 {
			if err := h.CreditsService.DeductCredits(user.ID, cost, turn.AssistantMessage.ID.String(), "Voice message"); err != nil {
				// Credit deduction failed - this is a billing issue that needs attention
				// The message was already processed, so we return it but log the error prominently
				log.Printf("CRITICAL: Failed to deduct credits for user %s, message %s: %v", user.ID, turn.AssistantMessage.ID, err)
				// Still return success since the message was processed - but this needs monitoring
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"userMessage":      turn.UserMessage,
		"assistantMessage": turn.AssistantMessage,
	})
}

// generateThreadName generates a thread name from AI response content (runs async)
func (h *ThreadHandler) generateThreadName(threadID uuid.UUID, aiResponse string) {
	// Check if thread already has a name
	thread, err := h.threadRepo.FindByID(h.exec, threadID)
	if err != nil {
		log.Printf("Error fetching thread for naming: %v", err)
		return
	}

	if thread.Name != nil {
		return // Already named
	}

	// Generate title from AI response
	title, err := h.OpenAIClient.GenerateTitle(aiResponse)
	if err != nil {
		log.Printf("Error generating thread title: %v", err)
		return
	}

	// Update thread name
	if err := h.threadRepo.UpdateName(h.exec, threadID, title); err != nil {
		log.Printf("Error updating thread name: %v", err)
	}
}

// UpdateThreadRequest represents the request body for updating a thread
type UpdateThreadRequest struct {
	Name *string `json:"name"`
}

// UpdateThread updates a thread's properties (rename)
func (h *ThreadHandler) UpdateThread(c *gin.Context) {
	user := middleware.MustGetUser(c)
	threadID := c.Param("id")

	parsedID, err := uuid.Parse(threadID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	thread, err := h.threadRepo.FindByIDAndUserID(h.exec, parsedID, user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch thread"})
		return
	}

	var req UpdateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		thread.Name = req.Name
	}

	if err := h.threadRepo.Save(h.exec, thread); err != nil {
		log.Printf("Error updating thread: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update thread"})
		return
	}

	c.JSON(http.StatusOK, thread)
}

// DeleteThread deletes a thread and all its messages
func (h *ThreadHandler) DeleteThread(c *gin.Context) {
	user := middleware.MustGetUser(c)
	threadID := c.Param("id")

	parsedID, err := uuid.Parse(threadID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	thread, err := h.threadRepo.FindByIDAndUserID(h.exec, parsedID, user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch thread"})
		return
	}

	// Delete thread (messages cascade delete via GORM constraint)
	if err := h.threadRepo.Delete(h.exec, thread); err != nil {
		log.Printf("Error deleting thread: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete thread"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thread deleted"})
}

// ArchiveThread sets the ArchivedAt timestamp on a thread
func (h *ThreadHandler) ArchiveThread(c *gin.Context) {
	user := middleware.MustGetUser(c)
	threadID := c.Param("id")

	parsedID, err := uuid.Parse(threadID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	thread, err := h.threadRepo.FindByIDAndUserID(h.exec, parsedID, user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch thread"})
		return
	}

	now := time.Now()
	thread.ArchivedAt = &now

	if err := h.threadRepo.Save(h.exec, thread); err != nil {
		log.Printf("Error archiving thread: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive thread"})
		return
	}

	c.JSON(http.StatusOK, thread)
}

// UnarchiveThread clears the ArchivedAt timestamp on a thread
func (h *ThreadHandler) UnarchiveThread(c *gin.Context) {
	user := middleware.MustGetUser(c)
	threadID := c.Param("id")

	parsedID, err := uuid.Parse(threadID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	thread, err := h.threadRepo.FindByIDAndUserID(h.exec, parsedID, user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch thread"})
		return
	}

	thread.ArchivedAt = nil

	if err := h.threadRepo.Save(h.exec, thread); err != nil {
		log.Printf("Error unarchiving thread: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unarchive thread"})
		return
	}

	c.JSON(http.StatusOK, thread)
}
