package client

import (
	"context"
	"io"
	"time"
)

// MLClient handles pronunciation analysis via the ML service.
type MLClient interface {
	AnalyzePronunciation(ctx context.Context, audioURL, expectedText, language string) (*PronunciationResponse, error)
}

// WhisperClient handles speech-to-text transcription.
type WhisperClient interface {
	TranscribeFromURL(ctx context.Context, audioURL string) (*TranscriptionResult, error)
}

// TTSClient handles text-to-speech synthesis.
type TTSClient interface {
	Synthesize(ctx context.Context, text string) (*TTSResult, error)
	SynthesizeWithOptions(ctx context.Context, text string, exaggeration float64, format string) (*TTSResult, error)
}

// OpenAIClient handles LLM generation via OpenAI.
type OpenAIClient interface {
	Generate(messages []ConversationMessage) (string, error)
	GenerateTitle(content string) (string, error)
}

// StorageClient handles object storage operations.
type StorageClient interface {
	UploadAudio(ctx context.Context, file io.Reader, key string, contentType string) (string, error)
	GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error)
	DeleteAudio(ctx context.Context, key string) error
	EnsureBucketExists(ctx context.Context) error
}

// ConversationMessage represents a chat message for LLM generation.
type ConversationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// TranscriptionResult is the result from speech-to-text.
type TranscriptionResult struct {
	Text     string
	Language string
	Duration float64
}

// TTSResult is the result from text-to-speech.
type TTSResult struct {
	AudioBytes []byte
	Duration   float64
}

// Pronunciation analysis types

// PronunciationResponse is the full response from pronunciation analysis.
type PronunciationResponse struct {
	Status   string                 `json:"status"`
	Analysis *PronunciationAnalysis `json:"analysis,omitempty"`
	Error    *PronunciationError    `json:"error,omitempty"`
}

// PronunciationAnalysis contains the detailed pronunciation analysis.
type PronunciationAnalysis struct {
	AudioIPA          string          `json:"audio_ipa"`
	ExpectedIPA       string          `json:"expected_ipa"`
	PhonemeCount      int             `json:"phoneme_count"`
	MatchCount        int             `json:"match_count"`
	SubstitutionCount int             `json:"substitution_count"`
	DeletionCount     int             `json:"deletion_count"`
	InsertionCount    int             `json:"insertion_count"`
	PhonemeDetails    []PhonemeDetail `json:"phoneme_details"`
	AudioQuality      *AudioQuality   `json:"audio_quality,omitempty"`
	ProcessingTimeMs  int64           `json:"processing_time_ms"`
}

// PhonemeDetail represents a single phoneme comparison.
type PhonemeDetail struct {
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Type     string `json:"type"`
	Position int    `json:"position"`
}

// AudioQuality contains audio quality metrics.
type AudioQuality struct {
	QualityScore    float64  `json:"quality_score"`
	SNRDB           float64  `json:"snr_db"`
	DurationSeconds float64  `json:"duration_seconds"`
	Warnings        []string `json:"warnings"`
}

// PronunciationError represents an error from pronunciation analysis.
type PronunciationError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}
