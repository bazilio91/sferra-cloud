package image

import (
	"context"
	"fmt"
	"github.com/bazilio91/sferra-cloud/pkg/services/storage"
	"io"
	"strings"
	"time"
)

// Service handles image-related operations
type Service struct {
	s3Client *storage.S3Client
}

// NewService creates a new image service
func NewService(s3Client *storage.S3Client) *Service {
	return &Service{
		s3Client: s3Client,
	}
}

// UploadResult represents the result of an image upload
type UploadResult struct {
	ID  string
	URL string
}

// UploadTaskImage handles image upload with metadata
func (s *Service) UploadTaskImage(ctx context.Context, clientID uint, taskID uint, filename string, reader io.Reader) (*UploadResult, error) {
	// Generate a unique key for the image using client ID, task ID and original filename
	key := fmt.Sprintf("images/%d/%d/%s", clientID, taskID, filename)

	// UploadTaskImage to S3
	if err := s.s3Client.UploadImage(ctx, key, reader); err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	// Generate a presigned URL that expires in 1 hour
	url, err := s.s3Client.GetPresignedURL(ctx, key, time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate URL: %w", err)
	}

	return &UploadResult{
		ID:  key,
		URL: url,
	}, nil
}

// GetResult represents the result of an image retrieval
type GetResult struct {
	Content     io.ReadCloser
	ContentType string
}

// GetTaskImage retrieves an image and verifies ownership using path
func (s *Service) GetTaskImage(ctx context.Context, clientID uint, taskID string, imageID string) (*GetResult, error) {
	imagePath := fmt.Sprintf("images/%d/%s/%s", clientID, taskID, imageID)
	content, contentType, err := s.s3Client.GetObject(ctx, imagePath)
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, fmt.Errorf("image not found")
		}
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return &GetResult{
		Content:     content,
		ContentType: contentType,
	}, nil
}
