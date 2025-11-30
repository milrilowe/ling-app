package handlers

import (
	"log"
	"net/http"
	"time"

	"ling-app/api/internal/db"
	"ling-app/api/internal/models"
	"ling-app/api/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ThreadHandler struct {
	DB           *db.DB
	OpenAIClient *services.OpenAIClient
}

func NewThreadHandler(database *db.DB, openAIClient *services.OpenAIClient) *ThreadHandler {
	return &ThreadHandler{
		DB:           database,
		OpenAIClient: openAIClient,
	}
}

type CreateThreadRequest struct {
	InitialPrompt    string `json:"initialPrompt"`
	FirstUserMessage string `json:"firstUserMessage"`
}

type SendMessageRequest struct {
	Content string `json:"content"`
}

// GetThreads retrieves all threads, ordered by most recent
func (h *ThreadHandler) GetThreads(c *gin.Context) {
	var threads []models.Thread
	if err := h.DB.Order("created_at DESC").Find(&threads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch threads"})
		return
	}
	c.JSON(http.StatusOK, threads)
}

// CreateThread creates a new conversation thread
func (h *ThreadHandler) CreateThread(c *gin.Context) {
	var req CreateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	thread := models.Thread{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
	}

	// Create thread
	if err := h.DB.Create(&thread).Error; err != nil {
		log.Printf("Error creating thread: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create thread"})
		return
	}

	// Add initial AI prompt message
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
		conversationHistory := []services.ConversationMessage{
			{Role: "assistant", Content: req.InitialPrompt},
			{Role: "user", Content: req.FirstUserMessage},
		}

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

// GetThread retrieves a thread with all messages
func (h *ThreadHandler) GetThread(c *gin.Context) {
	threadID := c.Param("id")

	parsedID, err := uuid.Parse(threadID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	var thread models.Thread
	if err := h.DB.Preload("Messages").First(&thread, parsedID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	c.JSON(http.StatusOK, thread)
}

// SendMessage sends a user message and gets AI response
func (h *ThreadHandler) SendMessage(c *gin.Context) {
	threadID := c.Param("id")
	parsedID, err := uuid.Parse(threadID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if thread exists
	var thread models.Thread
	if err := h.DB.First(&thread, parsedID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	// Save user message
	userMessage := models.Message{
		ID:        uuid.New(),
		ThreadID:  parsedID,
		Role:      "user",
		Content:   req.Content,
		Timestamp: time.Now(),
	}

	if err := h.DB.Create(&userMessage).Error; err != nil {
		log.Printf("Error creating user message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	// Get conversation history
	var messages []models.Message
	if err := h.DB.Where("thread_id = ?", parsedID).Order("timestamp ASC").Find(&messages).Error; err != nil {
		log.Printf("Error fetching messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	// Convert to ML format
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

	// Save AI response
	responseMessage := models.Message{
		ID:        uuid.New(),
		ThreadID:  parsedID,
		Role:      "assistant",
		Content:   aiResponse,
		Timestamp: time.Now(),
	}

	if err := h.DB.Create(&responseMessage).Error; err != nil {
		log.Printf("Error creating AI response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	c.JSON(http.StatusOK, responseMessage)
}
