package services

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type StorageService struct {
	client         *s3.Client
	internalClient *s3.Client // Client configured with internal Docker endpoint for service-to-service calls
	bucket         string
}

// NewStorageService creates a new storage service for S3/MinIO
func NewStorageService(endpoint, internalEndpoint, accessKey, secretKey, bucket, region string) (*StorageService, error) {
	// Create custom resolver for MinIO endpoint (external - for frontend access)
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, regionID string, options ...interface{}) (aws.Endpoint, error) {
		if endpoint != "" {
			return aws.Endpoint{
				URL:               endpoint,
				SigningRegion:     region,
				HostnameImmutable: true,
			}, nil
		}
		// Return empty endpoint to use default AWS resolver
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	// Load AWS config with custom credentials and endpoint
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client (for external/frontend access)
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // Required for MinIO
	})

	// Create internal client (for Docker-to-Docker service access, e.g., MFA container)
	var internalClient *s3.Client
	if internalEndpoint != "" && internalEndpoint != endpoint {
		internalResolver := aws.EndpointResolverWithOptionsFunc(func(service, regionID string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               internalEndpoint,
				SigningRegion:     region,
				HostnameImmutable: true,
			}, nil
		})

		internalCfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithEndpointResolverWithOptions(internalResolver),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load internal AWS config: %w", err)
		}

		internalClient = s3.NewFromConfig(internalCfg, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	return &StorageService{
		client:         client,
		internalClient: internalClient,
		bucket:         bucket,
	}, nil
}

// UploadAudio uploads an audio file to S3/MinIO
func (s *StorageService) UploadAudio(ctx context.Context, file io.Reader, key string, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return the key (path) to the file
	return key, nil
}

// GetPresignedURL generates a presigned URL for audio access (external endpoint)
func (s *StorageService) GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// GetInternalPresignedURL generates a presigned URL using the internal Docker endpoint
// This is for service-to-service calls where the calling service is in Docker (e.g., MFA container)
func (s *StorageService) GetInternalPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	// Use internal client if available, otherwise fall back to regular client
	client := s.internalClient
	if client == nil {
		client = s.client
	}

	presignClient := s3.NewPresignClient(client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))

	if err != nil {
		return "", fmt.Errorf("failed to generate internal presigned URL: %w", err)
	}

	return request.URL, nil
}

// DeleteAudio deletes an audio file from S3/MinIO
func (s *StorageService) DeleteAudio(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// EnsureBucketExists creates the bucket if it doesn't exist
func (s *StorageService) EnsureBucketExists(ctx context.Context) error {
	// Check if bucket exists
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})

	if err == nil {
		// Bucket already exists
		return nil
	}

	// Create bucket
	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}
