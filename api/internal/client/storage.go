package client

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

// storageClient implements StorageClient using S3/MinIO.
type storageClient struct {
	client *s3.Client
	bucket string
}

// NewStorageClient creates a new storage client for S3/MinIO.
// In production (isProduction=true), it uses IAM role credentials via the default AWS credential chain.
// In development (isProduction=false), it uses static credentials for local MinIO.
func NewStorageClient(endpoint, accessKey, secretKey, bucket, region string, isProduction bool) (StorageClient, error) {
	var cfg aws.Config
	var err error

	if isProduction {
		// Production: Use IAM role credentials via default AWS credential chain
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
		)
	} else {
		// Local development: Use static credentials for MinIO
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, regionID string, options ...interface{}) (aws.Endpoint, error) {
			if endpoint != "" {
				return aws.Endpoint{
					URL:               endpoint,
					SigningRegion:     region,
					HostnameImmutable: true,
				}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})

		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithEndpointResolverWithOptions(customResolver),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// Only use path style for MinIO (local development)
		o.UsePathStyle = !isProduction
	})

	return &storageClient{
		client: client,
		bucket: bucket,
	}, nil
}

// UploadAudio uploads an audio file to S3/MinIO.
func (s *storageClient) UploadAudio(ctx context.Context, file io.Reader, key string, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return key, nil
}

// GetPresignedURL generates a presigned URL for audio access.
func (s *storageClient) GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
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

// DeleteAudio deletes an audio file from S3/MinIO.
func (s *storageClient) DeleteAudio(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// EnsureBucketExists creates the bucket if it doesn't exist.
func (s *storageClient) EnsureBucketExists(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})

	if err == nil {
		return nil
	}

	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}
