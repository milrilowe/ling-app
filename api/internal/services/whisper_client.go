package services

import (
	"context"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

type WhisperClient struct {
	client *openai.Client
}

type TranscriptionResult struct {
	Text     string
	Language string
	Duration float64
}

// NewWhisperClient creates a new Whisper client for speech-to-text
func NewWhisperClient(apiKey string) *WhisperClient {
	client := openai.NewClient(apiKey)
	return &WhisperClient{
		client: client,
	}
}

// Transcribe transcribes audio file to text using OpenAI Whisper API
func (w *WhisperClient) Transcribe(ctx context.Context, audioFile io.Reader, filename string) (*TranscriptionResult, error) {
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: filename,
		Reader:   audioFile,
	}

	resp, err := w.client.CreateTranscription(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// OpenAI Whisper API doesn't return duration in the response
	// We'll need to calculate it separately if needed
	return &TranscriptionResult{
		Text:     resp.Text,
		Language: resp.Language,
		Duration: float64(resp.Duration),
	}, nil
}
