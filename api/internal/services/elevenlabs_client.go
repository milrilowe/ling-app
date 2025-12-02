package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ElevenLabsClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

type TTSResult struct {
	AudioBytes []byte
	Duration   float64
}

// NewElevenLabsClient creates a new ElevenLabs client for text-to-speech
func NewElevenLabsClient(apiKey string) *ElevenLabsClient {
	return &ElevenLabsClient{
		apiKey:  apiKey,
		baseURL: "https://api.elevenlabs.io/v1",
		client:  &http.Client{},
	}
}

// TextToSpeech converts text to speech using ElevenLabs API
func (e *ElevenLabsClient) TextToSpeech(text string, voiceID string) (*TTSResult, error) {
	// Default voice ID (Rachel) if not provided
	if voiceID == "" {
		voiceID = "21m00Tcm4TlvDq8ikWAM"
	}

	// Prepare request body
	requestBody := map[string]interface{}{
		"text": text,
		"model_id": "eleven_multilingual_v2",
		"voice_settings": map[string]interface{}{
			"stability":        0.5,
			"similarity_boost": 0.75,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/text-to-speech/%s", e.baseURL, voiceID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", e.apiKey)

	// Send request
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ElevenLabs API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Read audio bytes
	audioBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Note: ElevenLabs doesn't provide duration in the response
	// We'll need to calculate it separately if needed
	return &TTSResult{
		AudioBytes: audioBytes,
		Duration:   0, // To be calculated if needed
	}, nil
}
