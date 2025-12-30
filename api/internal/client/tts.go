package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ttsClient implements TTSClient using HTTP calls to the ML service.
type ttsClient struct {
	mlServiceURL string
	httpClient   *http.Client
}

// NewTTSClient creates a new TTS client that uses the ML service.
func NewTTSClient(mlServiceURL string) TTSClient {
	return &ttsClient{
		mlServiceURL: mlServiceURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // 1 minute timeout for TTS
		},
	}
}

// synthesizeRequest is the request body for the ML service /api/v1/synthesize endpoint.
type synthesizeRequest struct {
	Text         string  `json:"text"`
	Exaggeration float64 `json:"exaggeration"`
	Format       string  `json:"format"`
}

// synthesizeResponse is the response from the ML service.
type synthesizeResponse struct {
	Status      string   `json:"status"`
	AudioBase64 *string  `json:"audio_base64,omitempty"`
	Duration    *float64 `json:"duration,omitempty"`
	Format      *string  `json:"format,omitempty"`
	Error       *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Synthesize generates speech from text using the ML service.
func (t *ttsClient) Synthesize(ctx context.Context, text string) (*TTSResult, error) {
	return t.SynthesizeWithOptions(ctx, text, 0.5, "mp3")
}

// SynthesizeWithOptions generates speech with custom options.
func (t *ttsClient) SynthesizeWithOptions(ctx context.Context, text string, exaggeration float64, format string) (*TTSResult, error) {
	reqBody := synthesizeRequest{
		Text:         text,
		Exaggeration: exaggeration,
		Format:       format,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/synthesize", t.mlServiceURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result synthesizeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "success" {
		errMsg := "unknown error"
		if result.Error != nil {
			errMsg = result.Error.Message
		}
		return nil, fmt.Errorf("synthesis failed: %s", errMsg)
	}

	if result.AudioBase64 == nil {
		return nil, fmt.Errorf("synthesis succeeded but no audio data returned")
	}

	// Decode base64 audio
	audioBytes, err := base64.StdEncoding.DecodeString(*result.AudioBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode audio data: %w", err)
	}

	duration := 0.0
	if result.Duration != nil {
		duration = *result.Duration
	}

	return &TTSResult{
		AudioBytes: audioBytes,
		Duration:   duration,
	}, nil
}
