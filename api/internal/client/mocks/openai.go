package mocks

import (
	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/client"
)

// MockOpenAIClient is a mock implementation of OpenAIClient for testing.
type MockOpenAIClient struct {
	mock.Mock
}

// Ensure MockOpenAIClient implements client.OpenAIClient.
var _ client.OpenAIClient = (*MockOpenAIClient)(nil)

func (m *MockOpenAIClient) Generate(messages []client.ConversationMessage) (string, error) {
	args := m.Called(messages)
	return args.String(0), args.Error(1)
}

func (m *MockOpenAIClient) GenerateTitle(content string) (string, error) {
	args := m.Called(content)
	return args.String(0), args.Error(1)
}
