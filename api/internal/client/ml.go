package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// mlClient implements MLClient using HTTP calls to the ML service.
type mlClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewMLClient creates a new ML client.
func NewMLClient(baseURL string, timeout time.Duration) MLClient {
	if timeout == 0 {
		timeout = 120 * time.Second // Default 2 minutes for pronunciation analysis
	}
	return &mlClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// pronunciationRequest is the request body for the ML service.
type pronunciationRequest struct {
	AudioURL     string `json:"audio_url"`
	ExpectedText string `json:"expected_text"`
	Language     string `json:"language"`
}

// AnalyzePronunciation calls the ML service to analyze pronunciation.
func (c *mlClient) AnalyzePronunciation(ctx context.Context, audioURL, expectedText, language string) (*PronunciationResponse, error) {
	if language == "" {
		language = "en-us"
	}

	reqBody := pronunciationRequest{
		AudioURL:     audioURL,
		ExpectedText: expectedText,
		Language:     language,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/analyze-pronunciation", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	// Decode response regardless of status code to check for structured errors
	var result PronunciationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for ML service errors in response body
	if result.Status == "error" && result.Error != nil {
		return nil, &MLServiceError{
			Code:      result.Error.Code,
			Message:   result.Error.Message,
			Retryable: result.Error.Retryable,
		}
	}

	// Check HTTP status code for non-200 responses without structured errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned status %d", resp.StatusCode)
	}

	return &result, nil
}
