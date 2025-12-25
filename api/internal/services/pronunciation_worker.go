package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"ling-app/api/internal/db"
	"ling-app/api/internal/models"

	"github.com/google/uuid"
)

// PronunciationWorker handles async pronunciation analysis
type PronunciationWorker struct {
	DB       *db.DB
	MLClient *MLClient
	Storage  *StorageService
}

// NewPronunciationWorker creates a new pronunciation worker
func NewPronunciationWorker(database *db.DB, mlClient *MLClient, storage *StorageService) *PronunciationWorker {
	return &PronunciationWorker{
		DB:       database,
		MLClient: mlClient,
		Storage:  storage,
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

	// Convert analysis to JSON
	analysisJSON, err := json.Marshal(result.Analysis)
	if err != nil {
		log.Printf("[PronunciationWorker] Failed to marshal analysis: %v", err)
		w.markFailed(messageID, "JSON_ERROR", err.Error())
		return
	}

	// Update message with results
	now := time.Now()
	analysisStr := string(analysisJSON)
	err = w.DB.Model(&models.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"pronunciation_status":     "complete",
			"pronunciation_analysis":   analysisStr,
			"pronunciation_error":      nil,
			"pronunciation_updated_at": now,
		}).Error

	if err != nil {
		log.Printf("[PronunciationWorker] Failed to update message: %v", err)
		return
	}

	log.Printf("[PronunciationWorker] Analysis complete for message %s: %d/%d phonemes matched",
		messageID, result.Analysis.MatchCount, result.Analysis.PhonemeCount)
}

// markFailed updates the message with a failed status
func (w *PronunciationWorker) markFailed(messageID uuid.UUID, code, message string) {
	now := time.Now()
	errMsg := code + ": " + message
	err := w.DB.Model(&models.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"pronunciation_status":     "failed",
			"pronunciation_error":      errMsg,
			"pronunciation_updated_at": now,
		}).Error

	if err != nil {
		log.Printf("[PronunciationWorker] Failed to update message with error status: %v", err)
	}
}

// MarkPending marks a message as pending for pronunciation analysis
func (w *PronunciationWorker) MarkPending(messageID uuid.UUID) error {
	return w.DB.Model(&models.Message{}).
		Where("id = ?", messageID).
		Update("pronunciation_status", "pending").Error
}
