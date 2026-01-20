package mocks

import (
	"ling-app/api/internal/client"
	"ling-app/api/internal/services"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockPhonemeStatsProvider is a mock implementation of PhonemeStatsProvider interface
type MockPhonemeStatsProvider struct {
	mock.Mock
}

// GetUserStats mocks the GetUserStats method
func (m *MockPhonemeStatsProvider) GetUserStats(userID uuid.UUID) (*services.UserPhonemeStatsResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.UserPhonemeStatsResponse), args.Error(1)
}

// RecordPhonemeResults mocks the RecordPhonemeResults method
func (m *MockPhonemeStatsProvider) RecordPhonemeResults(userID uuid.UUID, phonemeDetails []client.PhonemeDetail) error {
	args := m.Called(userID, phonemeDetails)
	return args.Error(0)
}
