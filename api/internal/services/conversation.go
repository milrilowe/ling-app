package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"time"

	"ling-app/api/internal/client"
	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"

	"github.com/google/uuid"
)

// ConversationProcessor defines the interface for processing conversation messages
type ConversationProcessor interface {
	ProcessAudioMessage(ctx context.Context, threadID uuid.UUID, audioFile multipart.File, fileHeader *multipart.FileHeader) (*ConversationTurn, error)
}

// ConversationService handles audio message processing and AI conversation flow
type ConversationService struct {
	exec                repository.Executor
	messageRepo         repository.MessageRepository
	threadRepo          repository.ThreadRepository
	whisperClient       client.WhisperClient
	openAIClient        client.OpenAIClient
	ttsClient           client.TTSClient
	storage             client.StorageClient
	pronunciationWorker *PronunciationWorker
	maxAudioFileSize    int64
}

// ConversationTurn represents a complete user-assistant conversation exchange
type ConversationTurn struct {
	UserMessage      *models.Message `json:"userMessage"`
	AssistantMessage *models.Message `json:"assistantMessage"`
}

// NewConversationService creates a new conversation service
func NewConversationService(
	exec repository.Executor,
	messageRepo repository.MessageRepository,
	threadRepo repository.ThreadRepository,
	whisperClient client.WhisperClient,
	openAIClient client.OpenAIClient,
	ttsClient client.TTSClient,
	storage client.StorageClient,
	pronunciationWorker *PronunciationWorker,
	maxAudioFileSize int64,
) *ConversationService {
	return &ConversationService{
		exec:                exec,
		messageRepo:         messageRepo,
		threadRepo:          threadRepo,
		whisperClient:       whisperClient,
		openAIClient:        openAIClient,
		ttsClient:           ttsClient,
		storage:             storage,
		pronunciationWorker: pronunciationWorker,
		maxAudioFileSize:    maxAudioFileSize,
	}
}

// ProcessAudioMessage handles the complete flow of processing an audio message
// and generating an AI response with TTS audio
func (s *ConversationService) ProcessAudioMessage(
	ctx context.Context,
	threadID uuid.UUID,
	audioFile multipart.File,
	fileHeader *multipart.FileHeader,
) (*ConversationTurn, error) {
	// Validate file size
	if s.maxAudioFileSize > 0 && fileHeader.Size > s.maxAudioFileSize {
		return nil, fmt.Errorf("audio file too large: %d bytes (max: %d)", fileHeader.Size, s.maxAudioFileSize)
	}

	// Process user audio message
	userMessage, err := s.processUserAudio(ctx, threadID, audioFile, fileHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to process user audio: %w", err)
	}

	// Generate assistant response
	assistantMessage, err := s.generateAssistantResponse(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate assistant response: %w", err)
	}

	return &ConversationTurn{
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
	}, nil
}

// processUserAudio handles audio upload, transcription, and message creation
func (s *ConversationService) processUserAudio(
	ctx context.Context,
	threadID uuid.UUID,
	audioFile multipart.File,
	fileHeader *multipart.FileHeader,
) (*models.Message, error) {
	// Create user message ID
	userMessageID := uuid.New()

	// Upload user audio to storage
	userAudioKey := fmt.Sprintf("user/%s/%s.webm", threadID, userMessageID)
	_, err := s.storage.UploadAudio(ctx, audioFile, userAudioKey, "audio/webm")
	if err != nil {
		return nil, fmt.Errorf("failed to upload audio: %w", err)
	}

	// Get presigned URL for ML service to access the audio
	audioPresignedURL, err := s.storage.GetPresignedURL(ctx, userAudioKey, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to get presigned URL: %w", err)
	}

	// Transcribe audio
	transcription, err := s.whisperClient.TranscribeFromURL(ctx, audioPresignedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Validate audio duration
	if transcription.Duration < 1.0 {
		return nil, ErrAudioTooShort
	}
	if transcription.Duration > 30.0 {
		return nil, ErrAudioTooLong
	}

	// Save user message with audio (pronunciation analysis pending)
	userMessage := models.Message{
		ID:                   userMessageID,
		ThreadID:             threadID,
		Role:                 "user",
		Content:              transcription.Text,
		AudioURL:             &userAudioKey,
		AudioDurationSeconds: &transcription.Duration,
		HasAudio:             true,
		Timestamp:            time.Now(),
		PronunciationStatus:  "pending",
	}

	if err := s.messageRepo.Create(s.exec, &userMessage); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Spawn pronunciation analysis in background (non-blocking)
	if s.pronunciationWorker != nil {
		go s.pronunciationWorker.AnalyzeAsync(userMessageID, userAudioKey, transcription.Text, "en-us")
	}

	return &userMessage, nil
}

// generateAssistantResponse generates AI response with TTS audio
func (s *ConversationService) generateAssistantResponse(
	ctx context.Context,
	threadID uuid.UUID,
) (*models.Message, error) {
	// Get conversation history
	messages, err := s.messageRepo.FindByThreadID(s.exec, threadID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	// Convert to OpenAI format
	conversationHistory := make([]client.ConversationMessage, len(messages))
	for i, msg := range messages {
		conversationHistory[i] = client.ConversationMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Generate AI response
	aiResponse, err := s.openAIClient.Generate(conversationHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI response: %w", err)
	}

	assistantMessageID := uuid.New()

	// Try to generate TTS for AI response
	ttsResult, err := s.ttsClient.Synthesize(ctx, aiResponse)
	if err != nil {
		log.Printf("Error generating TTS: %v", err)
		// Continue without audio - save text-only response
		return s.createAssistantMessage(assistantMessageID, threadID, aiResponse, nil, nil, false)
	}

	// Upload TTS audio to storage
	assistantAudioKey := fmt.Sprintf("assistant/%s/%s.mp3", threadID, assistantMessageID)
	audioReader := bytes.NewReader(ttsResult.AudioBytes)
	_, err = s.storage.UploadAudio(ctx, audioReader, assistantAudioKey, "audio/mpeg")
	if err != nil {
		log.Printf("Error uploading TTS audio: %v", err)
		// Continue without audio
		return s.createAssistantMessage(assistantMessageID, threadID, aiResponse, nil, nil, false)
	}

	// Save AI response with audio
	ttsDuration := ttsResult.Duration
	return s.createAssistantMessage(assistantMessageID, threadID, aiResponse, &assistantAudioKey, &ttsDuration, true)
}

// createAssistantMessage creates and saves an assistant message
func (s *ConversationService) createAssistantMessage(
	messageID uuid.UUID,
	threadID uuid.UUID,
	content string,
	audioURL *string,
	audioDuration *float64,
	hasAudio bool,
) (*models.Message, error) {
	responseMessage := models.Message{
		ID:                   messageID,
		ThreadID:             threadID,
		Role:                 "assistant",
		Content:              content,
		AudioURL:             audioURL,
		AudioDurationSeconds: audioDuration,
		HasAudio:             hasAudio,
		Timestamp:            time.Now(),
	}

	if err := s.messageRepo.Create(s.exec, &responseMessage); err != nil {
		return nil, fmt.Errorf("failed to create AI response: %w", err)
	}

	return &responseMessage, nil
}
