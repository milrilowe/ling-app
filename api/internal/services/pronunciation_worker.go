package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"ling-app/api/internal/client"
	"ling-app/api/internal/db"
	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"

	"github.com/google/uuid"
)

// PronunciationWorker handles async pronunciation analysis
type PronunciationWorker struct {
	DB                  *db.DB
	exec                repository.Executor
	messageRepo         repository.MessageRepository
	threadRepo          repository.ThreadRepository
	MLClient            client.MLClient
	Storage             client.StorageClient
	PhonemeStatsService *PhonemeStatsService
}

// NewPronunciationWorker creates a new pronunciation worker
func NewPronunciationWorker(
	database *db.DB,
	mlClient client.MLClient,
	storage client.StorageClient,
	phonemeStatsService *PhonemeStatsService,
) *PronunciationWorker {
	return &PronunciationWorker{
		DB:                  database,
		exec:                database.DB,
		messageRepo:         repository.NewMessageRepository(),
		threadRepo:          repository.NewThreadRepository(),
		MLClient:            mlClient,
		Storage:             storage,
		PhonemeStatsService: phonemeStatsService,
	}
}

// NewPronunciationWorkerForTest creates a PronunciationWorker with injected dependencies for testing.
func NewPronunciationWorkerForTest(
	exec repository.Executor,
	messageRepo repository.MessageRepository,
	threadRepo repository.ThreadRepository,
	mlClient client.MLClient,
	storage client.StorageClient,
	phonemeStatsService *PhonemeStatsService,
) *PronunciationWorker {
	return &PronunciationWorker{
		DB:                  nil,
		exec:                exec,
		messageRepo:         messageRepo,
		threadRepo:          threadRepo,
		MLClient:            mlClient,
		Storage:             storage,
		PhonemeStatsService: phonemeStatsService,
	}
}

// AnalyzeAsync runs pronunciation analysis asynchronously
// This should be called from a goroutine so it doesn't block the HTTP response
func (w *PronunciationWorker) AnalyzeAsync(messageID uuid.UUID, audioKey, expectedText, language string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Printf("[PronunciationWorker] Starting analysis for message %s", messageID)

	// Generate presigned URL for the audio (1 hour expiration)
	presignedURL, err := w.Storage.GetPresignedURL(ctx, audioKey, 1*time.Hour)
	if err != nil {
		log.Printf("[PronunciationWorker] Failed to generate presigned URL: %v", err)
		w.markFailed(messageID, "PRESIGNED_URL_ERROR", err.Error())
		return
	}

	// Call ML service
	result, err := w.MLClient.AnalyzePronunciation(ctx, presignedURL, expectedText, language)
	if err != nil {
		log.Printf("[PronunciationWorker] ML service call failed: %v", err)
		w.markFailed(messageID, "ML_SERVICE_ERROR", err.Error())
		return
	}

	// Check if ML returned an error
	if result.Status == "error" {
		errMsg := "Unknown error"
		errCode := "UNKNOWN"
		if result.Error != nil {
			errMsg = result.Error.Message
			errCode = result.Error.Code
		}
		log.Printf("[PronunciationWorker] ML returned error: %s - %s", errCode, errMsg)
		w.markFailed(messageID, errCode, errMsg)
		return
	}

	// Success - store the analysis
	if result.Analysis == nil {
		log.Printf("[PronunciationWorker] ML returned success but no analysis data")
		w.markFailed(messageID, "NO_ANALYSIS", "ML service returned success but no analysis data")
		return
	}

	// Convert analysis to JSONMap for proper serialization
	analysisJSON, err := json.Marshal(result.Analysis)
	if err != nil {
		log.Printf("[PronunciationWorker] Failed to marshal analysis: %v", err)
		w.markFailed(messageID, "JSON_ERROR", err.Error())
		return
	}

	var analysisMap models.JSONMap
	if err := json.Unmarshal(analysisJSON, &analysisMap); err != nil {
		log.Printf("[PronunciationWorker] Failed to unmarshal analysis to map: %v", err)
		w.markFailed(messageID, "JSON_ERROR", err.Error())
		return
	}

	// Update message with results
	now := time.Now()
	if err := w.messageRepo.UpdatePronunciationAnalysis(w.exec, messageID, "complete", analysisMap, now); err != nil {
		log.Printf("[PronunciationWorker] Failed to update message: %v", err)
		return
	}

	log.Printf("[PronunciationWorker] Analysis complete for message %s: %d/%d phonemes matched",
		messageID, result.Analysis.MatchCount, result.Analysis.PhonemeCount)

	// Record phoneme stats for the user
	if w.PhonemeStatsService != nil && len(result.Analysis.PhonemeDetails) > 0 {
		// Get user ID from message -> thread -> user
		message, err := w.messageRepo.FindByID(w.exec, messageID)
		if err != nil {
			log.Printf("[PronunciationWorker] Failed to fetch message for phoneme stats: %v", err)
			return
		}

		thread, err := w.threadRepo.FindByID(w.exec, message.ThreadID)
		if err != nil {
			log.Printf("[PronunciationWorker] Failed to fetch thread for phoneme stats: %v", err)
			return
		}

		if err := w.PhonemeStatsService.RecordPhonemeResults(thread.UserID, result.Analysis.PhonemeDetails); err != nil {
			log.Printf("[PronunciationWorker] Failed to record phoneme stats: %v", err)
		} else {
			log.Printf("[PronunciationWorker] Recorded phoneme stats for user %s", thread.UserID)
		}
	}
}

// markFailed updates the message with a failed status
func (w *PronunciationWorker) markFailed(messageID uuid.UUID, code, message string) {
	now := time.Now()
	errMsg := code + ": " + message
	if err := w.messageRepo.UpdatePronunciationError(w.exec, messageID, "failed", errMsg, now); err != nil {
		log.Printf("[PronunciationWorker] Failed to update message with error status: %v", err)
	}
}

// MarkPending marks a message as pending for pronunciation analysis
func (w *PronunciationWorker) MarkPending(messageID uuid.UUID) error {
	return w.messageRepo.UpdatePronunciationStatus(w.exec, messageID, "pending")
}
