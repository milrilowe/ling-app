package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/client"
)

// MockWhisperClient is a mock implementation of WhisperClient for testing.
type MockWhisperClient struct {
	mock.Mock
}

// Ensure MockWhisperClient implements client.WhisperClient.
var _ client.WhisperClient = (*MockWhisperClient)(nil)

func (m *MockWhisperClient) TranscribeFromURL(ctx context.Context, audioURL string) (*client.TranscriptionResult, error) {
	args := m.Called(ctx, audioURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.TranscriptionResult), args.Error(1)
}
