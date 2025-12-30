package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/client"
)

// MockTTSClient is a mock implementation of TTSClient for testing.
type MockTTSClient struct {
	mock.Mock
}

// Ensure MockTTSClient implements client.TTSClient.
var _ client.TTSClient = (*MockTTSClient)(nil)

func (m *MockTTSClient) Synthesize(ctx context.Context, text string) (*client.TTSResult, error) {
	args := m.Called(ctx, text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.TTSResult), args.Error(1)
}

func (m *MockTTSClient) SynthesizeWithOptions(ctx context.Context, text string, exaggeration float64, format string) (*client.TTSResult, error) {
	args := m.Called(ctx, text, exaggeration, format)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.TTSResult), args.Error(1)
}
