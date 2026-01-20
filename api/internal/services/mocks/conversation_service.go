package mocks

import (
	"context"
	"mime/multipart"

	"ling-app/api/internal/services"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockConversationProcessor is a mock implementation of ConversationProcessor interface
type MockConversationProcessor struct {
	mock.Mock
}

// ProcessAudioMessage mocks the ProcessAudioMessage method
func (m *MockConversationProcessor) ProcessAudioMessage(
	ctx context.Context,
	threadID uuid.UUID,
	audioFile multipart.File,
	fileHeader *multipart.FileHeader,
) (*services.ConversationTurn, error) {
	args := m.Called(ctx, threadID, audioFile, fileHeader)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ConversationTurn), args.Error(1)
}
