package image

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bazilio91/sferra-cloud/pkg/services/storage"
	"testing"

	"github.com/bazilio91/sferra-cloud/pkg/config"
	"github.com/bazilio91/sferra-cloud/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Service, *storage.S3Client, func()) {
	ctx := context.Background()

	// Start MinIO container
	s3Container, err := testutils.StartS3Container(ctx)
	require.NoError(t, err)

	// Create test bucket
	err = s3Container.CreateBucket(ctx, "test-bucket")
	require.NoError(t, err)

	// Create S3 client with test config
	s3Config := s3Container.GetTestS3Config("test-bucket")
	cfg := &config.Config{
		S3Endpoint:        s3Config["endpoint"],
		S3Region:          s3Config["region"],
		S3Bucket:          s3Config["bucket"],
		S3AccessKeyID:     s3Config["accessKeyID"],
		S3SecretAccessKey: s3Config["secretAccessKey"],
	}
	s3Client := storage.NewS3Client(cfg)

	service := NewService(s3Client)

	cleanup := func() {
		err := s3Container.Container.Terminate(ctx)
		if err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return service, s3Client, cleanup
}

func TestImageService_Upload(t *testing.T) {
	service, _, cleanup := setupTest(t)
	defer cleanup()

	clientID := uint64(1)
	taskID := "123"
	filename := "test.jpg"
	content := []byte("test image data")

	result, err := service.UploadTaskImage(context.Background(), clientID, taskID, filename, bytes.NewReader(content))
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("images/%d/%s/%s", clientID, taskID, filename), result.ID)
	assert.Contains(t, result.URL, filename)
}

func TestImageService_Get(t *testing.T) {
	service, s3Client, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()
	clientID := uint64(1)
	taskID := "123"
	filename := "test.jpg"
	content := []byte("test image data")

	// UploadTaskImage directly to S3
	imagePath := fmt.Sprintf("images/%d/%s/%s", clientID, taskID, filename)
	err := s3Client.UploadImage(ctx, imagePath, bytes.NewReader(content))
	require.NoError(t, err)

	// GetTaskImage the image
	result, err := service.GetTaskImage(ctx, clientID, taskID, filename)
	require.NoError(t, err)

	// Read the content
	var buf bytes.Buffer
	_, err = buf.ReadFrom(result.Content)
	require.NoError(t, err)

	assert.Equal(t, content, buf.Bytes())
}

func TestImageService_GetInvalidPath(t *testing.T) {
	service, _, cleanup := setupTest(t)
	defer cleanup()

	// Test with invalid path format
	_, err := service.GetTaskImage(context.Background(), 1, "123", "invalid/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "image not found")
}
