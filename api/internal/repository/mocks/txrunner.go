package mocks

import (
	"database/sql"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockTxRunner is a mock implementation of auth.TxRunner for testing.
// It captures the transaction function and allows tests to control its behavior.
type MockTxRunner struct {
	mock.Mock
}

func (m *MockTxRunner) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	args := m.Called(fc, opts)
	return args.Error(0)
}

// RunTransaction is a helper that executes the captured transaction function
// with a mock gorm.DB (nil in tests, since repository mocks handle the logic).
// Call this when you want the transaction to actually execute.
func (m *MockTxRunner) RunTransaction(fc func(tx *gorm.DB) error) error {
	return fc(nil)
}
