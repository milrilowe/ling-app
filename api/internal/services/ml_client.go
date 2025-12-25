package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MLClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewMLClient(baseURL string, timeout time.Duration) *MLClient {
	if timeout == 0 {
		timeout = 120 * time.Second // Default 2 minutes for pronunciation analysis
	}
	return &MLClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type GenerateRequest struct {
	Messages []ConversationMessage `json:"messages"`
}

type ConversationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GenerateResponse struct {
	Content string `json:"content"`
}

// Pronunciation analysis types

type PronunciationRequest struct {
	AudioURL     string `json:"audio_url"`
	ExpectedText string `json:"expected_text"`
	Language     string `json:"language"`
}

type PhonemeDetail struct {
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Type     string `json:"type"`
	Position int    `json:"position"`
}

type AudioQuality struct {
	QualityScore    float64  `json:"quality_score"`
	SNRDB           float64  `json:"snr_db"`
	DurationSeconds float64  `json:"duration_seconds"`
	Warnings        []string `json:"warnings"`
}

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

type PronunciationError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

type PronunciationResponse struct {
	Status   string                 `json:"status"`
	Analysis *PronunciationAnalysis `json:"analysis,omitempty"`
	Error    *PronunciationError    `json:"error,omitempty"`
}

// Generate calls the ML service to generate an AI response
func (c *MLClient) Generate(messages []ConversationMessage) (string, error) {
	reqBody := GenerateRequest{
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := c.HTTPClient.Post(
		c.BaseURL+"/ml/generate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ML service returned status %d", resp.StatusCode)
	}

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Content, nil
}

// AnalyzePronunciation calls the ML service to analyze pronunciation
func (c *MLClient) AnalyzePronunciation(ctx context.Context, audioURL, expectedText, language string) (*PronunciationResponse, error) {
	if language == "" {
		language = "en-us"
	}

	reqBody := PronunciationRequest{
		AudioURL:     audioURL,
		ExpectedText: expectedText,
		Language:     language,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/v1/analyze-pronunciation", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned status %d", resp.StatusCode)
	}

	var result PronunciationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
