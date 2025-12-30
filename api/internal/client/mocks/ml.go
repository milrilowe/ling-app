package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/client"
)

// MockMLClient is a mock implementation of MLClient for testing.
type MockMLClient struct {
	mock.Mock
}

// Ensure MockMLClient implements client.MLClient.
var _ client.MLClient = (*MockMLClient)(nil)

func (m *MockMLClient) AnalyzePronunciation(ctx context.Context, audioURL, expectedText, language string) (*client.PronunciationResponse, error) {
	args := m.Called(ctx, audioURL, expectedText, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.PronunciationResponse), args.Error(1)
}
