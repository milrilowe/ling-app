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

// TTSAlignmentWorker handles async word timing extraction for TTS audio
type TTSAlignmentWorker struct {
	DB        *db.DB
	MFAClient *MFAClient
	Storage   *StorageService
}

// NewTTSAlignmentWorker creates a new TTS alignment worker
func NewTTSAlignmentWorker(database *db.DB, mfaClient *MFAClient, storage *StorageService) *TTSAlignmentWorker {
	return &TTSAlignmentWorker{
		DB:        database,
		MFAClient: mfaClient,
		Storage:   storage,
	}
}

// AlignAsync runs MFA alignment asynchronously on assistant TTS audio
// This should be called from a goroutine so it doesn't block the HTTP response
func (w *TTSAlignmentWorker) AlignAsync(messageID uuid.UUID, audioKey, transcript string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Printf("[TTSAlignmentWorker] Starting alignment for message %s", messageID)

	// Mark as pending
	if err := w.MarkPending(messageID); err != nil {
		log.Printf("[TTSAlignmentWorker] Failed to mark pending: %v", err)
		return
	}

	// Generate presigned URL for the audio using internal Docker endpoint (1 hour expiration)
	// This allows the MFA container to access MinIO via Docker networking (http://minio:9000)
	presignedURL, err := w.Storage.GetInternalPresignedURL(ctx, audioKey, 1*time.Hour)
	if err != nil {
		log.Printf("[TTSAlignmentWorker] Failed to generate presigned URL: %v", err)
		w.markFailed(messageID, "PRESIGNED_URL_ERROR: "+err.Error())
		return
	}
	log.Printf("[TTSAlignmentWorker] Generated internal URL for MFA: %s", presignedURL)

	// Call MFA service
	result, err := w.MFAClient.Align(ctx, presignedURL, transcript, "english_us_arpa")
	if err != nil {
		log.Printf("[TTSAlignmentWorker] MFA alignment failed: %v", err)
		w.markFailed(messageID, "MFA_ERROR: "+err.Error())
		return
	}

	// Convert word timings to JSONMap
	wordTimingsJSON, err := json.Marshal(result.Words)
	if err != nil {
		log.Printf("[TTSAlignmentWorker] Failed to marshal word timings: %v", err)
		w.markFailed(messageID, "JSON_ERROR: "+err.Error())
		return
	}

	var wordTimingsMap models.JSONMap
	if err := json.Unmarshal(wordTimingsJSON, &wordTimingsMap); err != nil {
		// If it's an array, wrap it
		var wordTimingsArray []interface{}
		if err := json.Unmarshal(wordTimingsJSON, &wordTimingsArray); err != nil {
			log.Printf("[TTSAlignmentWorker] Failed to unmarshal word timings: %v", err)
			w.markFailed(messageID, "JSON_ERROR: "+err.Error())
			return
		}
		// Store as {"words": [...]} for consistent access
		wordTimingsMap = models.JSONMap{"words": wordTimingsArray}
	}

	// Update message with word timings
	err = w.DB.Model(&models.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"word_timings":        wordTimingsMap,
			"word_timings_status": "complete",
			"word_timings_error":  nil,
		}).Error

	if err != nil {
		log.Printf("[TTSAlignmentWorker] Failed to update message: %v", err)
		return
	}

	log.Printf("[TTSAlignmentWorker] Alignment complete for message %s: %d words",
		messageID, len(result.Words))
}

// MarkPending marks a message as pending for word timing alignment
func (w *TTSAlignmentWorker) MarkPending(messageID uuid.UUID) error {
	return w.DB.Model(&models.Message{}).
		Where("id = ?", messageID).
		Update("word_timings_status", "pending").Error
}

// markFailed updates the message with a failed status
func (w *TTSAlignmentWorker) markFailed(messageID uuid.UUID, errMsg string) {
	err := w.DB.Model(&models.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"word_timings_status": "failed",
			"word_timings_error":  errMsg,
		}).Error

	if err != nil {
		log.Printf("[TTSAlignmentWorker] Failed to update message with error status: %v", err)
	}
}
