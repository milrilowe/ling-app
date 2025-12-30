package mocks

import (
	"context"
	"io"
	"time"

	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/client"
)

// MockStorageClient is a mock implementation of StorageClient for testing.
type MockStorageClient struct {
	mock.Mock
}

// Ensure MockStorageClient implements client.StorageClient.
var _ client.StorageClient = (*MockStorageClient)(nil)

func (m *MockStorageClient) UploadAudio(ctx context.Context, file io.Reader, key string, contentType string) (string, error) {
	args := m.Called(ctx, file, key, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockStorageClient) GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, key, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockStorageClient) DeleteAudio(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockStorageClient) EnsureBucketExists(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
