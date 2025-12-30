package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// whisperClient implements WhisperClient using HTTP calls to the ML service.
type whisperClient struct {
	mlServiceURL string
	httpClient   *http.Client
}

// NewWhisperClient creates a new Whisper client that uses the ML service.
func NewWhisperClient(mlServiceURL string) WhisperClient {
	return &whisperClient{
		mlServiceURL: mlServiceURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // 2 minute timeout for transcription
		},
	}
}

// transcribeRequest is the request body for the ML service /api/v1/transcribe endpoint.
type transcribeRequest struct {
	AudioURL string  `json:"audio_url"`
	Language *string `json:"language,omitempty"`
}

// transcribeResponse is the response from the ML service.
type transcribeResponse struct {
	Status   string   `json:"status"`
	Text     *string  `json:"text,omitempty"`
	Language *string  `json:"language,omitempty"`
	Duration *float64 `json:"duration,omitempty"`
	Error    *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// TranscribeFromURL transcribes audio from a presigned URL using the ML service.
func (w *whisperClient) TranscribeFromURL(ctx context.Context, audioURL string) (*TranscriptionResult, error) {
	reqBody := transcribeRequest{
		AudioURL: audioURL,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/transcribe", w.mlServiceURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
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

	var result transcribeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "success" {
		errMsg := "unknown error"
		if result.Error != nil {
			errMsg = result.Error.Message
		}
		return nil, fmt.Errorf("transcription failed: %s", errMsg)
	}

	text := ""
	if result.Text != nil {
		text = *result.Text
	}

	language := ""
	if result.Language != nil {
		language = *result.Language
	}

	duration := 0.0
	if result.Duration != nil {
		duration = *result.Duration
	}

	return &TranscriptionResult{
		Text:     text,
		Language: language,
		Duration: duration,
	}, nil
}
