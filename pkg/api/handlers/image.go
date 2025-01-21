package handlers

import (
	"fmt"
	"github.com/bazilio91/sferra-cloud/pkg/services/image"
	"github.com/bazilio91/sferra-cloud/pkg/services/storage"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	imageService *image.Service
}

func NewImageHandler(s3Client *storage.S3Client) *ImageHandler {
	return &ImageHandler{
		imageService: image.NewService(s3Client),
	}
}

// getClientIDFromContext extracts client ID from the JWT claims
func getClientIDFromContext(c *gin.Context) (uint, error) {
	userValue, exists := c.Get("user")
	if !exists {
		return 0, fmt.Errorf("user not found in context")
	}

	claims, ok := userValue.(*auth.Claims)
	if !ok {
		return 0, fmt.Errorf("invalid user claims type")
	}

	return uint(claims.ClientID), nil
}

// UploadImage godoc
// @Summary UploadTaskImage an image
// @Description UploadTaskImage an image to storage
// @Accept multipart/form-data
// @Produce json
// @Param task_id path uint true "Recognition Task ID"
// @Param image formData file true "Image file"
// @Success 200 {object} map[string]string
// @Router /recognition-tasks/{task_id}/images/upload [post]
func (h *ImageHandler) UploadImage(c *gin.Context) {
	clientID, err := getClientIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// GetTaskImage task ID from path
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	defer src.Close()

	result, err := h.imageService.UploadTaskImage(c.Request.Context(), clientID, uint(taskID), file.Filename, src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":  result.ID,
		"url": result.URL,
	})
}

// GetImage godoc
// @Summary GetTaskImage an image
// @Description GetTaskImage an image by ID
// @Produce image/jpeg,image/png,image/gif
// @Param task_id path uint true "Recognition Task ID"
// @Param image_id path string true "Image ID"
// @Success 200 {file} binary
// @Router /recognition-tasks/{task_id}/images/{image_id} [get]
func (h *ImageHandler) GetImage(c *gin.Context) {
	clientID, err := getClientIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	taskID := c.Param("task_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	filename := c.Param("image_id")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image ID is required"})
		return
	}

	// URL decode the filename
	decodedFilename, err := url.QueryUnescape(filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image path: invalid image path format"})
		return
	}

	// Check if the filename contains any path separators
	if strings.Contains(decodedFilename, "/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image path: invalid image path format"})
		return
	}

	imageData, err := h.imageService.GetTaskImage(c.Request.Context(), clientID, taskID, decodedFilename)
	if err != nil {
		if strings.Contains(err.Error(), "access denied: image belongs to different client") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "invalid image path") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "image not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer imageData.Content.Close()

	// Set content type header
	c.Header("Content-Type", imageData.ContentType)
	c.Header("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	c.Header("Content-Disposition", "inline")

	c.DataFromReader(http.StatusOK, -1, imageData.ContentType, imageData.Content, nil)
}
