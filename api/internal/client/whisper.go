package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// openAIWhisperClient uses OpenAI Whisper API.
type openAIWhisperClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewOpenAIWhisperClient creates a Whisper client using OpenAI API.
func NewOpenAIWhisperClient(apiKey string) WhisperClient {
	return &openAIWhisperClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (w *openAIWhisperClient) TranscribeFromURL(ctx context.Context, audioURL string) (*TranscriptionResult, error) {
	// Download audio from URL first
	audioResp, err := http.Get(audioURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio: %w", err)
	}
	defer audioResp.Body.Close()

	if audioResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download audio: status %d", audioResp.StatusCode)
	}

	audioBytes, err := io.ReadAll(audioResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio: %w", err)
	}

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add audio file
	part, err := writer.CreateFormFile("file", "audio.webm")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(audioBytes); err != nil {
		return nil, fmt.Errorf("failed to write audio: %w", err)
	}

	// Add model field
	if err := writer.WriteField("model", "whisper-1"); err != nil {
		return nil, fmt.Errorf("failed to write model field: %w", err)
	}

	// Add response format for duration
	if err := writer.WriteField("response_format", "verbose_json"); err != nil {
		return nil, fmt.Errorf("failed to write response_format field: %w", err)
	}

	writer.Close()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/audio/transcriptions", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+w.apiKey)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI Whisper API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI Whisper API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Text     string  `json:"text"`
		Language string  `json:"language"`
		Duration float64 `json:"duration"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &TranscriptionResult{
		Text:     result.Text,
		Language: result.Language,
		Duration: result.Duration,
	}, nil
}

// mlWhisperClient uses the ML service (faster-whisper).
type mlWhisperClient struct {
	mlServiceURL string
	httpClient   *http.Client
}

// NewMLWhisperClient creates a Whisper client using the ML service.
func NewMLWhisperClient(mlServiceURL string) WhisperClient {
	return &mlWhisperClient{
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
func (w *mlWhisperClient) TranscribeFromURL(ctx context.Context, audioURL string) (*TranscriptionResult, error) {
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
