package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MFAClient handles communication with the MFA alignment service
type MFAClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewMFAClient creates a new MFA client
func NewMFAClient(baseURL string, timeout time.Duration) *MFAClient {
	if timeout == 0 {
		timeout = 60 * time.Second // Default 1 minute for alignment
	}
	return &MFAClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// AlignRequest is the request body for the alignment endpoint
type AlignRequest struct {
	AudioURL   string `json:"audio_url"`
	Transcript string `json:"transcript"`
	Language   string `json:"language"`
}

// WordTiming represents a word with its timing information
type WordTiming struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// PhoneTiming represents a phoneme with timing and dual format (ARPAbet + IPA)
type PhoneTiming struct {
	ARPAbet string  `json:"arpabet"`
	IPA     string  `json:"ipa"`
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
}

// AlignResponse is the response from the alignment endpoint
type AlignResponse struct {
	Words    []WordTiming  `json:"words"`
	Phones   []PhoneTiming `json:"phones"`
	Duration float64       `json:"duration"`
}

// Align calls the MFA service to align audio with transcript
// Returns word and phoneme-level timing information
func (c *MFAClient) Align(ctx context.Context, audioURL, transcript, language string) (*AlignResponse, error) {
	if language == "" {
		language = "english_us_arpa"
	}

	reqBody := AlignRequest{
		AudioURL:   audioURL,
		Transcript: transcript,
		Language:   language,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/v1/align", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call MFA service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MFA service returned status %d", resp.StatusCode)
	}

	var result AlignResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// HealthCheck verifies the MFA service is running
func (c *MFAClient) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call MFA service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MFA service returned status %d", resp.StatusCode)
	}

	return nil
}
