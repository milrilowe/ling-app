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

// openAITTSClient uses OpenAI TTS API.
type openAITTSClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewOpenAITTSClient creates a TTS client using OpenAI TTS API.
func NewOpenAITTSClient(apiKey string) TTSClient {
	return &openAITTSClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (t *openAITTSClient) Synthesize(ctx context.Context, text string) (*TTSResult, error) {
	return t.SynthesizeWithOptions(ctx, text, 1.0, "mp3")
}

func (t *openAITTSClient) SynthesizeWithOptions(ctx context.Context, text string, exaggeration float64, format string) (*TTSResult, error) {
	reqBody := map[string]interface{}{
		"model":           "tts-1",
		"input":           text,
		"voice":           "alloy", // Options: alloy, echo, fable, onyx, nova, shimmer
		"response_format": format,
		"speed":           1.0,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/audio/speech", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.apiKey)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI TTS API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI TTS API returned status %d: %s", resp.StatusCode, string(body))
	}

	audioBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio response: %w", err)
	}

	// Estimate duration (~150 words/min, ~5 chars/word)
	estimatedDuration := float64(len(text)) / (5.0 * 150.0 / 60.0)

	return &TTSResult{
		AudioBytes: audioBytes,
		Duration:   estimatedDuration,
	}, nil
}

// mlTTSClient uses the ML service (Chatterbox).
type mlTTSClient struct {
	mlServiceURL string
	httpClient   *http.Client
}

// NewMLTTSClient creates a TTS client using the ML service (Chatterbox).
func NewMLTTSClient(mlServiceURL string) TTSClient {
	return &mlTTSClient{
		mlServiceURL: mlServiceURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (t *mlTTSClient) Synthesize(ctx context.Context, text string) (*TTSResult, error) {
	return t.SynthesizeWithOptions(ctx, text, 0.5, "mp3")
}

func (t *mlTTSClient) SynthesizeWithOptions(ctx context.Context, text string, exaggeration float64, format string) (*TTSResult, error) {
	reqBody := map[string]interface{}{
		"text":         text,
		"exaggeration": exaggeration,
		"format":       format,
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

	var result struct {
		Status      string   `json:"status"`
		AudioBase64 *string  `json:"audio_base64,omitempty"`
		Duration    *float64 `json:"duration,omitempty"`
		Error       *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

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
