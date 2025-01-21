package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bazilio91/sferra-cloud/pkg/config"
)

type S3Client struct {
	client *s3.Client
	bucket string
}

func NewS3Client(cfg *config.Config) *S3Client {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				URL:           cfg.S3Endpoint,
				SigningRegion: cfg.S3Region,
				Source:        aws.EndpointSourceCustom,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg := aws.Config{
		Credentials:                 credentials.NewStaticCredentialsProvider(cfg.S3AccessKeyID, cfg.S3SecretAccessKey, ""),
		Region:                      cfg.S3Region,
		EndpointResolverWithOptions: customResolver,
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Client{
		client: client,
		bucket: cfg.S3Bucket,
	}
}

func (s *S3Client) UploadImage(ctx context.Context, key string, reader io.Reader) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return fmt.Errorf("failed to upload image: %w", err)
	}
	return nil
}

func (s *S3Client) GetImage(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	return result.Body, nil
}

func (s *S3Client) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(po *s3.PresignOptions) {
		po.Expires = expires
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// GetObject retrieves an object and its content type
func (s *S3Client) GetObject(ctx context.Context, key string) (io.ReadCloser, string, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object: %w", err)
	}

	contentType := "application/octet-stream"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	return result.Body, contentType, nil
}

// GetObjectMetadata retrieves metadata for an object
func (s *S3Client) GetObjectMetadata(ctx context.Context, key string) (map[string]string, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.HeadObject(ctx, input)
	if err != nil {
		return nil, err
	}

	return result.Metadata, nil
}

// UploadImageWithMetadata uploads an image with metadata
func (s *S3Client) UploadImageWithMetadata(ctx context.Context, key string, reader io.Reader, metadata map[string]string) error {
	input := &s3.PutObjectInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(key),
		Body:     reader,
		Metadata: metadata,
	}

	// Set content type if provided in metadata
	if contentType, ok := metadata["content-type"]; ok {
		input.ContentType = aws.String(contentType)
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload image with metadata: %w", err)
	}
	return nil
}
