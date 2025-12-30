package services

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/client"
	clientmocks "ling-app/api/internal/client/mocks"
	"ling-app/api/internal/models"
	repomocks "ling-app/api/internal/repository/mocks"
)

func TestPronunciationWorker_AnalyzeAsync_Success(t *testing.T) {
	messageID := uuid.New()
	userID := uuid.New()
	threadID := uuid.New()
	audioKey := "audio/test.wav"
	expectedText := "hello"
	language := "en"

	// Setup mocks
	messageRepo := new(repomocks.MockMessageRepository)
	threadRepo := new(repomocks.MockThreadRepository)
	storageClient := new(clientmocks.MockStorageClient)
	mlClient := new(clientmocks.MockMLClient)
	phonemeStatsRepo := new(repomocks.MockPhonemeStatsRepository)
	phonemeSubsRepo := new(repomocks.MockPhonemeSubstitutionRepository)

	phonemeStatsService := NewPhonemeStatsServiceForTest(nil, phonemeStatsRepo, phonemeSubsRepo)

	// Storage returns presigned URL
	storageClient.On("GetPresignedURL", mock.Anything, audioKey, time.Hour).
		Return("https://presigned.url/test.wav", nil)

	// ML returns successful analysis
	mlClient.On("AnalyzePronunciation", mock.Anything, "https://presigned.url/test.wav", expectedText, language).
		Return(&client.PronunciationResponse{
			Status: "success",
			Analysis: &client.PronunciationAnalysis{
				PhonemeCount: 5,
				MatchCount:   4,
				PhonemeDetails: []client.PhonemeDetail{
					{Expected: "h", Actual: "h", Type: "match"},
					{Expected: "ɛ", Actual: "ɛ", Type: "match"},
					{Expected: "l", Actual: "l", Type: "match"},
					{Expected: "oʊ", Actual: "oʊ", Type: "match"},
				},
			},
		}, nil)

	// Message repo updates with analysis
	messageRepo.On("UpdatePronunciationAnalysis", mock.Anything, messageID, "complete", mock.AnythingOfType("models.JSONMap"), mock.AnythingOfType("time.Time")).
		Return(nil)

	// For phoneme stats, we need to get message and thread
	messageRepo.On("FindByID", mock.Anything, messageID).
		Return(&models.Message{ID: messageID, ThreadID: threadID}, nil)

	threadRepo.On("FindByID", mock.Anything, threadID).
		Return(&models.Thread{ID: threadID, UserID: userID}, nil)

	// Phoneme stats recording (for match phonemes)
	phonemeStatsRepo.On("Upsert", mock.Anything, mock.Anything).Return(nil)

	worker := NewPronunciationWorkerForTest(nil, messageRepo, threadRepo, mlClient, storageClient, phonemeStatsService)
	worker.AnalyzeAsync(messageID, audioKey, expectedText, language)

	storageClient.AssertExpectations(t)
	mlClient.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
	threadRepo.AssertExpectations(t)
}

func TestPronunciationWorker_AnalyzeAsync_PresignedURLError(t *testing.T) {
	messageID := uuid.New()

	messageRepo := new(repomocks.MockMessageRepository)
	threadRepo := new(repomocks.MockThreadRepository)
	storageClient := new(clientmocks.MockStorageClient)
	mlClient := new(clientmocks.MockMLClient)

	storageClient.On("GetPresignedURL", mock.Anything, "audio/test.wav", time.Hour).
		Return("", errors.New("storage error"))

	messageRepo.On("UpdatePronunciationError", mock.Anything, messageID, "failed", "PRESIGNED_URL_ERROR: storage error", mock.AnythingOfType("time.Time")).
		Return(nil)

	worker := NewPronunciationWorkerForTest(nil, messageRepo, threadRepo, mlClient, storageClient, nil)
	worker.AnalyzeAsync(messageID, "audio/test.wav", "hello", "en")

	storageClient.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
	// ML client should not be called
	mlClient.AssertNotCalled(t, "AnalyzePronunciation")
}

func TestPronunciationWorker_AnalyzeAsync_MLServiceError(t *testing.T) {
	messageID := uuid.New()

	messageRepo := new(repomocks.MockMessageRepository)
	threadRepo := new(repomocks.MockThreadRepository)
	storageClient := new(clientmocks.MockStorageClient)
	mlClient := new(clientmocks.MockMLClient)

	storageClient.On("GetPresignedURL", mock.Anything, "audio/test.wav", time.Hour).
		Return("https://presigned.url/test.wav", nil)

	mlClient.On("AnalyzePronunciation", mock.Anything, "https://presigned.url/test.wav", "hello", "en").
		Return(nil, errors.New("ML service unavailable"))

	messageRepo.On("UpdatePronunciationError", mock.Anything, messageID, "failed", "ML_SERVICE_ERROR: ML service unavailable", mock.AnythingOfType("time.Time")).
		Return(nil)

	worker := NewPronunciationWorkerForTest(nil, messageRepo, threadRepo, mlClient, storageClient, nil)
	worker.AnalyzeAsync(messageID, "audio/test.wav", "hello", "en")

	storageClient.AssertExpectations(t)
	mlClient.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
}

func TestPronunciationWorker_AnalyzeAsync_MLReturnsError(t *testing.T) {
	messageID := uuid.New()

	messageRepo := new(repomocks.MockMessageRepository)
	threadRepo := new(repomocks.MockThreadRepository)
	storageClient := new(clientmocks.MockStorageClient)
	mlClient := new(clientmocks.MockMLClient)

	storageClient.On("GetPresignedURL", mock.Anything, "audio/test.wav", time.Hour).
		Return("https://presigned.url/test.wav", nil)

	mlClient.On("AnalyzePronunciation", mock.Anything, "https://presigned.url/test.wav", "hello", "en").
		Return(&client.PronunciationResponse{
			Status: "error",
			Error: &client.PronunciationError{
				Code:    "AUDIO_TOO_SHORT",
				Message: "Audio is too short for analysis",
			},
		}, nil)

	messageRepo.On("UpdatePronunciationError", mock.Anything, messageID, "failed", "AUDIO_TOO_SHORT: Audio is too short for analysis", mock.AnythingOfType("time.Time")).
		Return(nil)

	worker := NewPronunciationWorkerForTest(nil, messageRepo, threadRepo, mlClient, storageClient, nil)
	worker.AnalyzeAsync(messageID, "audio/test.wav", "hello", "en")

	storageClient.AssertExpectations(t)
	mlClient.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
}

func TestPronunciationWorker_AnalyzeAsync_NoAnalysisData(t *testing.T) {
	messageID := uuid.New()

	messageRepo := new(repomocks.MockMessageRepository)
	threadRepo := new(repomocks.MockThreadRepository)
	storageClient := new(clientmocks.MockStorageClient)
	mlClient := new(clientmocks.MockMLClient)

	storageClient.On("GetPresignedURL", mock.Anything, "audio/test.wav", time.Hour).
		Return("https://presigned.url/test.wav", nil)

	mlClient.On("AnalyzePronunciation", mock.Anything, "https://presigned.url/test.wav", "hello", "en").
		Return(&client.PronunciationResponse{
			Status:   "success",
			Analysis: nil, // No analysis data
		}, nil)

	messageRepo.On("UpdatePronunciationError", mock.Anything, messageID, "failed", "NO_ANALYSIS: ML service returned success but no analysis data", mock.AnythingOfType("time.Time")).
		Return(nil)

	worker := NewPronunciationWorkerForTest(nil, messageRepo, threadRepo, mlClient, storageClient, nil)
	worker.AnalyzeAsync(messageID, "audio/test.wav", "hello", "en")

	storageClient.AssertExpectations(t)
	mlClient.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
}

func TestPronunciationWorker_MarkPending(t *testing.T) {
	messageID := uuid.New()

	t.Run("successfully marks message as pending", func(t *testing.T) {
		messageRepo := new(repomocks.MockMessageRepository)
		threadRepo := new(repomocks.MockThreadRepository)

		messageRepo.On("UpdatePronunciationStatus", mock.Anything, messageID, "pending").Return(nil)

		worker := NewPronunciationWorkerForTest(nil, messageRepo, threadRepo, nil, nil, nil)
		err := worker.MarkPending(messageID)

		assert.NoError(t, err)
		messageRepo.AssertExpectations(t)
	})

	t.Run("returns error when update fails", func(t *testing.T) {
		messageRepo := new(repomocks.MockMessageRepository)
		threadRepo := new(repomocks.MockThreadRepository)

		dbError := errors.New("database error")
		messageRepo.On("UpdatePronunciationStatus", mock.Anything, messageID, "pending").Return(dbError)

		worker := NewPronunciationWorkerForTest(nil, messageRepo, threadRepo, nil, nil, nil)
		err := worker.MarkPending(messageID)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
	})
}

func TestPronunciationWorker_AnalyzeAsync_WithSubstitutions(t *testing.T) {
	messageID := uuid.New()
	userID := uuid.New()
	threadID := uuid.New()

	messageRepo := new(repomocks.MockMessageRepository)
	threadRepo := new(repomocks.MockThreadRepository)
	storageClient := new(clientmocks.MockStorageClient)
	mlClient := new(clientmocks.MockMLClient)
	phonemeStatsRepo := new(repomocks.MockPhonemeStatsRepository)
	phonemeSubsRepo := new(repomocks.MockPhonemeSubstitutionRepository)

	phonemeStatsService := NewPhonemeStatsServiceForTest(nil, phonemeStatsRepo, phonemeSubsRepo)

	storageClient.On("GetPresignedURL", mock.Anything, "audio/test.wav", time.Hour).
		Return("https://presigned.url/test.wav", nil)

	// ML returns analysis with substitutions
	mlClient.On("AnalyzePronunciation", mock.Anything, "https://presigned.url/test.wav", "think", "en").
		Return(&client.PronunciationResponse{
			Status: "success",
			Analysis: &client.PronunciationAnalysis{
				PhonemeCount:      4,
				MatchCount:        3,
				SubstitutionCount: 1,
				PhonemeDetails: []client.PhonemeDetail{
					{Expected: "θ", Actual: "f", Type: "substitute"},
					{Expected: "ɪ", Actual: "ɪ", Type: "match"},
					{Expected: "ŋ", Actual: "ŋ", Type: "match"},
					{Expected: "k", Actual: "k", Type: "match"},
				},
			},
		}, nil)

	messageRepo.On("UpdatePronunciationAnalysis", mock.Anything, messageID, "complete", mock.AnythingOfType("models.JSONMap"), mock.AnythingOfType("time.Time")).
		Return(nil)

	messageRepo.On("FindByID", mock.Anything, messageID).
		Return(&models.Message{ID: messageID, ThreadID: threadID}, nil)

	threadRepo.On("FindByID", mock.Anything, threadID).
		Return(&models.Thread{ID: threadID, UserID: userID}, nil)

	// Stats and substitution recording
	phonemeStatsRepo.On("Upsert", mock.Anything, mock.Anything).Return(nil)
	phonemeSubsRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *models.PhonemeSubstitution) bool {
		return s.ExpectedPhoneme == "θ" && s.ActualPhoneme == "f"
	})).Return(nil)

	worker := NewPronunciationWorkerForTest(nil, messageRepo, threadRepo, mlClient, storageClient, phonemeStatsService)
	worker.AnalyzeAsync(messageID, "audio/test.wav", "think", "en")

	storageClient.AssertExpectations(t)
	mlClient.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
	threadRepo.AssertExpectations(t)
}
