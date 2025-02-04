package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aws/smithy-go/ptr"
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"github.com/google/uuid"

	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/gin-gonic/gin"
)

// DataRecognitionTaskListResponse represents a paginated list response
type DataRecognitionTaskListResponse struct {
	TotalCount int64                       `json:"total_count"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	Results    []proto.DataRecognitionTask `json:"results"`
}

// CreateDataRecognitionTask godoc
// @Summary Create DataRecognitionTask
// @Description Create a new DataRecognitionTask
// @Tags recognition_tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body proto.DataRecognitionTask true "DataRecognitionTask data"
// @Success 201 {object} proto.DataRecognitionTask
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/recognition_tasks [post]
func CreateDataRecognitionTask(c *gin.Context) {
	userClaims := c.MustGet("claims").(*auth.Claims)

	var request proto.DataRecognitionTask
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// GetTaskImage client
	var clientORM proto.ClientORM
	if err := db.DB.First(&clientORM, "id = ?", userClaims.ClientID).Error; err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "client not found"})
		return
	}

	// Generate UUID for the new task
	request.Id = uuid.New().String()
	request.Status = proto.Status_STATUS_IMAGES_PENDING

	// Convert to ORM model
	ormObj, err := request.ToORM(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Set client and timestamps
	ormObj.Client = &clientORM
	ormObj.CreatedAt = ptr.Time(time.Now())
	ormObj.UpdatedAt = ptr.Time(time.Now())

	// Create the task
	if err := db.DB.Create(&ormObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Convert back to proto
	response, err := ormObj.ToPB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetDataRecognitionTask godoc
// @Summary GetTaskImage DataRecognitionTask
// @Description GetTaskImage a DataRecognitionTask by ID
// @Tags recognition_tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "DataRecognitionTask ID"
// @Success 200 {object} proto.DataRecognitionTask
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/recognition_tasks/{id} [get]
func GetDataRecognitionTask(c *gin.Context) {
	userClaims := c.MustGet("claims").(*auth.Claims)
	id := c.Param("id")

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid task ID format: must be a valid UUID"})
		return
	}

	var ormObj proto.DataRecognitionTaskORM
	if err := db.DB.First(&ormObj, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "task not found"})
		return
	}

	// Check client access
	if ormObj.Client == nil || ormObj.Client.Id != uint64(userClaims.ClientID) {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "access denied"})
		return
	}

	// Convert to proto
	response, err := ormObj.ToPB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateDataRecognitionTask godoc
// @Summary Update UpdateDataRecognitionTask
// @Description Update a UpdateDataRecognitionTask by ID
// @Tags recognition_tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "UpdateDataRecognitionTask ID"
// @Param data body proto.DataRecognitionTask true "Updated DataRecognitionTask"
// @Success 200 {object} proto.DataRecognitionTask
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/recognition_tasks/{id} [put]
func UpdateDataRecognitionTask(c *gin.Context) {
	userClaims := c.MustGet("claims").(*auth.Claims)
	id := c.Param("id")

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid task ID format: must be a valid UUID"})
		return
	}

	// GetTaskImage existing task
	var existingORM proto.DataRecognitionTaskORM
	if err := db.DB.First(&existingORM, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "task not found"})
		return
	}

	// Check client access
	if existingORM.Client.Id != uint64(userClaims.ClientID) {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "access denied"})
		return
	}

	// Parse update request
	var updateRequest proto.DataRecognitionTask
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Convert update to ORM
	updateORM, err := updateRequest.ToORM(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Preserve immutable fields
	updateORM.Id = existingORM.Id
	updateORM.Client = existingORM.Client
	updateORM.CreatedAt = existingORM.CreatedAt
	updateORM.UpdatedAt = ptr.Time(time.Now())

	// Save updates
	if err := db.DB.Save(&updateORM).Error; err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Convert back to proto
	response, err := updateORM.ToPB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteDataRecognitionTask godoc
// @Summary Delete DataRecognitionTask
// @Description Delete a DataRecognitionTask by ID
// @Tags recognition_tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "DataRecognitionTask ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/recognition_tasks/{id} [delete]
func DeleteDataRecognitionTask(c *gin.Context) {
	userClaims := c.MustGet("claims").(*auth.Claims)
	id := c.Param("id")

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid task ID format: must be a valid UUID"})
		return
	}

	var ormObj proto.DataRecognitionTaskORM
	if err := db.DB.Preload("Client").First(&ormObj, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "task not found"})
		return
	}

	// Check client access
	if ormObj.Client == nil || ormObj.Client.Id != uint64(userClaims.ClientID) {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "access denied"})
		return
	}

	if err := db.DB.Delete(&ormObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "task deleted successfully"})
}

// ListDataRecognitionTask godoc
// @Summary List DataRecognitionTask
// @Description List DataRecognitionTasks for the authenticated client
// @Tags recognition_tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} DataRecognitionTaskListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/recognition_tasks [get]
func ListDataRecognitionTask(c *gin.Context) {
	userClaims := c.MustGet("claims").(*auth.Claims)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	query := db.DB.Model(&proto.DataRecognitionTaskORM{}).
		Joins("JOIN clients ON clients.id = data_recognition_tasks.client_id").
		Where("clients.id = ?", userClaims.ClientID)

	if status != "" {
		query = query.Where("data_recognition_tasks.status = ?", status)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	var ormResults []proto.DataRecognitionTaskORM
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&ormResults).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Convert ORM results to proto
	results := make([]proto.DataRecognitionTask, 0, len(ormResults))
	for _, orm := range ormResults {
		proto, err := orm.ToPB(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		results = append(results, proto)
	}

	c.JSON(http.StatusOK, DataRecognitionTaskListResponse{
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		Results:    results,
	})
}
