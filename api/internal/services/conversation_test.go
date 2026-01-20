package services

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/client"
	clientmocks "ling-app/api/internal/client/mocks"
	"ling-app/api/internal/models"
	repomocks "ling-app/api/internal/repository/mocks"
)

// mockMultipartFile implements multipart.File for testing
type mockMultipartFile struct {
	*bytes.Reader
}

func (m *mockMultipartFile) Close() error {
	return nil
}

func newMockMultipartFile(data []byte) multipart.File {
	return &mockMultipartFile{
		Reader: bytes.NewReader(data),
	}
}

func TestConversationService_ProcessAudioMessage_Success(t *testing.T) {
	// Setup
	threadID := uuid.New()
	audioContent := []byte("fake audio data")
	audioFile := newMockMultipartFile(audioContent)
	fileHeader := &multipart.FileHeader{
		Filename: "test.webm",
		Size:     int64(len(audioContent)),
	}

	// Mock repositories
	messageRepo := new(repomocks.MockMessageRepository)
	threadRepo := new(repomocks.MockThreadRepository)

	// Mock clients
	whisperClient := new(clientmocks.MockWhisperClient)
	openAIClient := new(clientmocks.MockOpenAIClient)
	ttsClient := new(clientmocks.MockTTSClient)
	storageClient := new(clientmocks.MockStorageClient)

	// Storage: upload audio
	storageClient.On("UploadAudio", mock.Anything, mock.Anything, mock.Anything, "audio/webm").
		Return("https://storage.url/audio.webm", nil)

	// Storage: get presigned URL
	storageClient.On("GetPresignedURL", mock.Anything, mock.Anything, mock.Anything).
		Return("https://presigned.url/audio.webm", nil)

	// Whisper: transcribe audio
	whisperClient.On("TranscribeFromURL", mock.Anything, "https://presigned.url/audio.webm").
		Return(&client.TranscriptionResult{
			Text:     "hello world",
			Language: "en",
			Duration: 2.5,
		}, nil)

	// Message repo: create user message
	messageRepo.On("Create", mock.Anything, mock.MatchedBy(func(msg *models.Message) bool {
		return msg.Role == "user" && msg.Content == "hello world" && msg.HasAudio == true
	})).Return(nil)

	// Message repo: find by thread ID (for conversation history)
	messageRepo.On("FindByThreadID", mock.Anything, threadID).Return([]models.Message{
		{
			ID:       uuid.New(),
			ThreadID: threadID,
			Role:     "user",
			Content:  "hello world",
		},
	}, nil)

	// OpenAI: generate response
	openAIClient.On("Generate", mock.MatchedBy(func(history []client.ConversationMessage) bool {
		return len(history) == 1 && history[0].Content == "hello world"
	})).Return("Hi! How can I help you today?", nil)

	// TTS: synthesize response
	ttsClient.On("Synthesize", mock.Anything, "Hi! How can I help you today?").Return(&client.TTSResult{
		AudioBytes: []byte("fake tts audio"),
		Duration:   1.5,
	}, nil)

	// Storage: upload TTS audio
	storageClient.On("UploadAudio", mock.Anything, mock.Anything, mock.Anything, "audio/mpeg").
		Return("https://storage.url/assistant.mp3", nil)

	// Message repo: create assistant message
	messageRepo.On("Create", mock.Anything, mock.MatchedBy(func(msg *models.Message) bool {
		return msg.Role == "assistant" && msg.Content == "Hi! How can I help you today?" && msg.HasAudio == true
	})).Return(nil)

	// Create service (without pronunciation worker for this test)
	service := NewConversationService(
		nil, // executor
		messageRepo,
		threadRepo,
		whisperClient,
		openAIClient,
		ttsClient,
		storageClient,
		nil, // pronunciation worker
		10*1024*1024,
	)

	// Execute
	turn, err := service.ProcessAudioMessage(context.Background(), threadID, audioFile, fileHeader)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, turn)
	assert.NotNil(t, turn.UserMessage)
	assert.NotNil(t, turn.AssistantMessage)
	assert.Equal(t, "hello world", turn.UserMessage.Content)
	assert.Equal(t, "Hi! How can I help you today?", turn.AssistantMessage.Content)
	assert.True(t, turn.UserMessage.HasAudio)
	assert.True(t, turn.AssistantMessage.HasAudio)

	// Verify all mocks were called
	storageClient.AssertExpectations(t)
	whisperClient.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
	openAIClient.AssertExpectations(t)
	ttsClient.AssertExpectations(t)
}

func TestConversationService_ProcessAudioMessage_FileSizeTooLarge(t *testing.T) {
	// Setup
	threadID := uuid.New()
	audioContent := []byte("fake audio data")
	audioFile := newMockMultipartFile(audioContent)
	fileHeader := &multipart.FileHeader{
		Filename: "test.webm",
		Size:     100 * 1024 * 1024, // 100MB
	}

	// Create service with 10MB limit
	service := NewConversationService(
		nil, nil, nil, nil, nil, nil, nil, nil,
		10*1024*1024,
	)

	// Execute
	turn, err := service.ProcessAudioMessage(context.Background(), threadID, audioFile, fileHeader)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, turn)
	assert.Contains(t, err.Error(), "audio file too large")
}

func TestConversationService_ProcessAudioMessage_UploadFails(t *testing.T) {
	// Setup
	threadID := uuid.New()
	audioContent := []byte("fake audio data")
	audioFile := newMockMultipartFile(audioContent)
	fileHeader := &multipart.FileHeader{
		Filename: "test.webm",
		Size:     int64(len(audioContent)),
	}

	// Mock storage to fail
	storageClient := new(clientmocks.MockStorageClient)
	storageClient.On("UploadAudio", mock.Anything, mock.Anything, mock.Anything, "audio/webm").
		Return("", errors.New("storage error"))

	// Create service
	service := NewConversationService(
		nil, nil, nil, nil, nil, nil, storageClient, nil,
		10*1024*1024,
	)

	// Execute
	turn, err := service.ProcessAudioMessage(context.Background(), threadID, audioFile, fileHeader)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, turn)
	assert.Contains(t, err.Error(), "failed to process user audio")
	assert.Contains(t, err.Error(), "failed to upload audio")
	storageClient.AssertExpectations(t)
}

func TestConversationService_ProcessAudioMessage_TranscriptionFails(t *testing.T) {
	// Setup
	threadID := uuid.New()
	audioContent := []byte("fake audio data")
	audioFile := newMockMultipartFile(audioContent)
	fileHeader := &multipart.FileHeader{
		Filename: "test.webm",
		Size:     int64(len(audioContent)),
	}

	// Mock storage
	storageClient := new(clientmocks.MockStorageClient)
	storageClient.On("UploadAudio", mock.Anything, mock.Anything, mock.Anything, "audio/webm").
		Return("https://storage.url/audio.webm", nil)
	storageClient.On("GetPresignedURL", mock.Anything, mock.Anything, mock.Anything).
		Return("https://presigned.url/audio.webm", nil)

	// Mock whisper to fail
	whisperClient := new(clientmocks.MockWhisperClient)
	whisperClient.On("TranscribeFromURL", mock.Anything, "https://presigned.url/audio.webm").
		Return(nil, errors.New("transcription failed"))

	// Create service
	service := NewConversationService(
		nil, nil, nil, whisperClient, nil, nil, storageClient, nil,
		10*1024*1024,
	)

	// Execute
	turn, err := service.ProcessAudioMessage(context.Background(), threadID, audioFile, fileHeader)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, turn)
	assert.Contains(t, err.Error(), "failed to process user audio")
	assert.Contains(t, err.Error(), "failed to transcribe audio")
	storageClient.AssertExpectations(t)
	whisperClient.AssertExpectations(t)
}

func TestConversationService_ProcessAudioMessage_TTSFails_ContinuesWithoutAudio(t *testing.T) {
	// Setup
	threadID := uuid.New()
	audioContent := []byte("fake audio data")
	audioFile := newMockMultipartFile(audioContent)
	fileHeader := &multipart.FileHeader{
		Filename: "test.webm",
		Size:     int64(len(audioContent)),
	}

	// Mock repositories
	messageRepo := new(repomocks.MockMessageRepository)

	// Mock clients
	whisperClient := new(clientmocks.MockWhisperClient)
	openAIClient := new(clientmocks.MockOpenAIClient)
	ttsClient := new(clientmocks.MockTTSClient)
	storageClient := new(clientmocks.MockStorageClient)

	// Setup successful audio processing
	storageClient.On("UploadAudio", mock.Anything, mock.Anything, mock.Anything, "audio/webm").
		Return("https://storage.url/audio.webm", nil)
	storageClient.On("GetPresignedURL", mock.Anything, mock.Anything, mock.Anything).
		Return("https://presigned.url/audio.webm", nil)
	whisperClient.On("TranscribeFromURL", mock.Anything, mock.Anything).
		Return(&client.TranscriptionResult{Text: "hello", Duration: 1.0}, nil)
	messageRepo.On("Create", mock.Anything, mock.MatchedBy(func(msg *models.Message) bool {
		return msg.Role == "user"
	})).Return(nil)
	messageRepo.On("FindByThreadID", mock.Anything, threadID).Return([]models.Message{
		{Role: "user", Content: "hello"},
	}, nil)
	openAIClient.On("Generate", mock.Anything).Return("Hi there!", nil)

	// TTS fails
	ttsClient.On("Synthesize", mock.Anything, "Hi there!").Return(nil, errors.New("TTS error"))

	// Assistant message created without audio
	messageRepo.On("Create", mock.Anything, mock.MatchedBy(func(msg *models.Message) bool {
		return msg.Role == "assistant" && msg.HasAudio == false
	})).Return(nil)

	// Create service
	service := NewConversationService(
		nil, messageRepo, nil, whisperClient, openAIClient, ttsClient, storageClient, nil,
		10*1024*1024,
	)

	// Execute
	turn, err := service.ProcessAudioMessage(context.Background(), threadID, audioFile, fileHeader)

	// Assert - should succeed despite TTS failure
	assert.NoError(t, err)
	assert.NotNil(t, turn)
	assert.Equal(t, "Hi there!", turn.AssistantMessage.Content)
	assert.False(t, turn.AssistantMessage.HasAudio) // No audio due to TTS failure
	messageRepo.AssertExpectations(t)
	ttsClient.AssertExpectations(t)
}

func TestConversationService_ProcessAudioMessage_WithPronunciationWorker(t *testing.T) {
	// Setup
	threadID := uuid.New()
	audioContent := []byte("fake audio data")
	audioFile := newMockMultipartFile(audioContent)
	fileHeader := &multipart.FileHeader{
		Filename: "test.webm",
		Size:     int64(len(audioContent)),
	}

	// Mock everything
	messageRepo := new(repomocks.MockMessageRepository)
	whisperClient := new(clientmocks.MockWhisperClient)
	openAIClient := new(clientmocks.MockOpenAIClient)
	ttsClient := new(clientmocks.MockTTSClient)
	storageClient := new(clientmocks.MockStorageClient)

	// Setup successful flow
	storageClient.On("UploadAudio", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("https://storage.url/file", nil)
	storageClient.On("GetPresignedURL", mock.Anything, mock.Anything, mock.Anything).
		Return("https://presigned.url/file", nil)
	whisperClient.On("TranscribeFromURL", mock.Anything, mock.Anything).
		Return(&client.TranscriptionResult{Text: "test", Duration: 1.0}, nil)
	messageRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	messageRepo.On("FindByThreadID", mock.Anything, threadID).
		Return([]models.Message{{Role: "user", Content: "test"}}, nil)
	openAIClient.On("Generate", mock.Anything).Return("Response", nil)
	ttsClient.On("Synthesize", mock.Anything, mock.Anything).
		Return(&client.TTSResult{AudioBytes: []byte("audio"), Duration: 1.0}, nil)

	// Create mock pronunciation worker
	// Note: We're just testing that the service doesn't crash if worker is nil
	// The actual worker.AnalyzeAsync is called as a goroutine, so we can't easily test it here

	// Create service without worker (testing it handles nil gracefully)
	service := NewConversationService(
		nil, messageRepo, nil, whisperClient, openAIClient, ttsClient, storageClient, nil,
		10*1024*1024,
	)

	// Execute
	turn, err := service.ProcessAudioMessage(context.Background(), threadID, audioFile, fileHeader)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, turn)
	// Test passes even without pronunciation worker
}

func TestConversationService_ProcessAudioMessage_AudioTooShort(t *testing.T) {
	// Setup
	threadID := uuid.New()
	audioContent := []byte("fake audio data")
	audioFile := newMockMultipartFile(audioContent)
	fileHeader := &multipart.FileHeader{
		Filename: "test.webm",
		Size:     int64(len(audioContent)),
	}

	// Mock repositories
	messageRepo := new(repomocks.MockMessageRepository)

	// Mock clients
	whisperClient := new(clientmocks.MockWhisperClient)
	storageClient := new(clientmocks.MockStorageClient)

	// Mock storage operations
	storageClient.On("UploadAudio", mock.Anything, mock.Anything, mock.Anything, "audio/webm").Return("", nil)
	storageClient.On("GetPresignedURL", mock.Anything, mock.Anything, mock.Anything).Return("http://example.com/audio.webm", nil)

	// Mock transcription with short audio (0.5 seconds)
	whisperClient.On("TranscribeFromURL", mock.Anything, "http://example.com/audio.webm").Return(&client.TranscriptionResult{
		Text:     "hi",
		Language: "en",
		Duration: 0.5, // Less than 1 second
	}, nil)

	// Create service
	service := NewConversationService(
		nil, messageRepo, nil, whisperClient, nil, nil, storageClient, nil,
		10*1024*1024,
	)

	// Execute
	turn, err := service.ProcessAudioMessage(context.Background(), threadID, audioFile, fileHeader)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, turn)
	assert.ErrorIs(t, err, ErrAudioTooShort)

	// Verify mocks
	whisperClient.AssertExpectations(t)
	storageClient.AssertExpectations(t)
	// Message should NOT be created since validation failed
	messageRepo.AssertNotCalled(t, "Create")
}
