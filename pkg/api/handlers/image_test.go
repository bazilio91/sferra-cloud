package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/bazilio91/sferra-cloud/pkg/services/storage"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/bazilio91/sferra-cloud/pkg/config"
	"github.com/bazilio91/sferra-cloud/pkg/testutils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupImageTest(t *testing.T) (*gin.Engine, *storage.S3Client, func()) {
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

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	r := gin.New()

	imageHandler := NewImageHandler(s3Client)

	// Add authentication middleware
	authMiddleware := func(c *gin.Context) {
		claims := &auth.Claims{
			UserID:   1,
			ClientID: 1,
		}
		c.Set("user", claims)
		c.Next()
	}

	// Add routes
	r.POST("/recognition-tasks/:task_id/images", authMiddleware, imageHandler.UploadImage)
	r.GET("/recognition-tasks/:task_id/images/:image_id", authMiddleware, imageHandler.GetImage)
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image path: invalid image path format"})
	})

	cleanup := func() {
		err := s3Container.Container.Terminate(ctx)
		if err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return r, s3Client, cleanup
}

func TestUploadImage(t *testing.T) {
	router, _, cleanup := setupImageTest(t)
	defer cleanup()

	// Create a test image
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	part, err := writer.CreateFormFile("image", "test.jpg")
	require.NoError(t, err)

	imageData := []byte("fake image content")
	_, err = part.Write(imageData)
	require.NoError(t, err)
	writer.Close()

	// Create request
	req := httptest.NewRequest("POST", "/recognition-tasks/123/images", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "id")
	assert.Contains(t, response, "url")
	assert.Contains(t, response["id"], "images/1/123/test.jpg")
	assert.Contains(t, response["url"], "test.jpg")
}

func TestUploadImageInvalidTaskID(t *testing.T) {
	router, _, cleanup := setupImageTest(t)
	defer cleanup()

	// Create a test image
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	part, err := writer.CreateFormFile("image", "test.jpg")
	require.NoError(t, err)

	imageData := []byte("fake image content")
	_, err = part.Write(imageData)
	require.NoError(t, err)
	writer.Close()

	// Create request with invalid task ID
	req := httptest.NewRequest("POST", "/recognition-tasks/invalid/images", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid task ID")
}

func TestGetImage(t *testing.T) {
	router, s3Client, cleanup := setupImageTest(t)
	defer cleanup()

	ctx := context.Background()
	filename := "test.jpg"
	imageData := []byte("test image data")

	// UploadTaskImage directly to S3
	imageKey := "images/1/123/" + filename
	err := s3Client.UploadImage(ctx, imageKey, bytes.NewReader(imageData))
	require.NoError(t, err)

	// Create request
	req := httptest.NewRequest("GET", "/recognition-tasks/123/images/"+filename, nil)

	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "public, max-age=31536000", w.Header().Get("Cache-Control"))
	assert.Equal(t, "inline", w.Header().Get("Content-Disposition"))
	assert.Equal(t, imageData, w.Body.Bytes())
}

func TestGetImageUnauthorized(t *testing.T) {
	router, s3Client, cleanup := setupImageTest(t)
	defer cleanup()

	ctx := context.Background()
	filename := "test.jpg"
	imageData := []byte("test image data")

	// UploadTaskImage image for client 2
	imageKey := "images/2/123/" + filename
	err := s3Client.UploadImage(ctx, imageKey, bytes.NewReader(imageData))
	require.NoError(t, err)

	// Try to access the image as client 1 (which should fail)
	req := httptest.NewRequest("GET", "/recognition-tasks/123/images/"+filename, nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "image not found")
}

func TestGetImageInvalidPath(t *testing.T) {
	router, _, cleanup := setupImageTest(t)
	defer cleanup()

	// Test with invalid path format by including a slash in the filename
	req := httptest.NewRequest("GET", "/recognition-tasks/123/images/invalid%2Fpath.jpg", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid image path: invalid image path format")
}

func TestGetImageInvalidTaskID(t *testing.T) {
	router, _, cleanup := setupImageTest(t)
	defer cleanup()

	// Test with invalid task ID
	req := httptest.NewRequest("GET", "/recognition-tasks/invalid/images/test.jpg", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "image not found")
}
