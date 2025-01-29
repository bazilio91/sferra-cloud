package admin

import (
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func ListRecognitionTasks(c *gin.Context) {
	var tasks []proto.DataRecognitionTaskORM
	query := db.DB.Order("created_at DESC")

	// Apply filters
	if id := c.Query("id"); id != "" {
		query = query.Where("id = ?", id)
	}
	if clientID := c.Query("client_id"); clientID != "" {
		query = query.Where("client_id = ?", clientID)
	}

	// Get clients for the filter dropdown
	var clients []proto.ClientORM
	if err := db.DB.Find(&clients).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "recognition_task/tasks.html", gin.H{
			"Error": "Failed to fetch clients",
		})
		return
	}

	if err := query.Preload("Client").Find(&tasks).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "recognition_task/tasks.html", gin.H{
			"Error": "Failed to fetch tasks",
		})
		return
	}

	c.HTML(http.StatusOK, "recognition_task/tasks.html", gin.H{
		"Tasks":   tasks,
		"Clients": clients,
		"Filters": gin.H{
			"ID":       c.Query("id"),
			"ClientID": c.Query("client_id"),
		},
	})
}

func EditRecognitionTask(c *gin.Context) {
	id := c.Param("id")
	var task proto.DataRecognitionTaskORM
	if err := db.DB.Preload("Client").First(&task, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.HTML(http.StatusInternalServerError, "recognition_task/edit.html", gin.H{
			"Error": "Failed to fetch task",
		})
		return
	}

	c.HTML(http.StatusOK, "recognition_task/edit.html", gin.H{
		"Task": task,
		"Statuses": []gin.H{
			{"Value": int32(proto.Status_STATUS_CREATED), "Label": "Created"},
			{"Value": int32(proto.Status_STATUS_READY_FOR_PROCESSING), "Label": "Ready for Processing"},
			{"Value": int32(proto.Status_STATUS_IMAGES_PENDING), "Label": "Images Pending"},
			{"Value": int32(proto.Status_STATUS_IMAGES_PROCESSING), "Label": "Images Processing"},
			{"Value": int32(proto.Status_STATUS_IMAGES_COMPLETED), "Label": "Images Completed"},
			{"Value": int32(proto.Status_STATUS_IMAGES_FAILED_QUOTA), "Label": "Images Failed (Quota)"},
			{"Value": int32(proto.Status_STATUS_IMAGES_FAILED_PROCESSING), "Label": "Images Failed (Processing)"},
			{"Value": int32(proto.Status_STATUS_IMAGES_FAILED_TIMEOUT), "Label": "Images Failed (Timeout)"},
			{"Value": int32(proto.Status_STATUS_RECOGNITION_PENDING), "Label": "Recognition Pending"},
			{"Value": int32(proto.Status_STATUS_RECOGNITION_PROCESSING), "Label": "Recognition Processing"},
			{"Value": int32(proto.Status_STATUS_RECOGNITION_COMPLETED), "Label": "Recognition Completed"},
			{"Value": int32(proto.Status_STATUS_RECOGNITION_FAILED_QUOTA), "Label": "Recognition Failed (Quota)"},
			{"Value": int32(proto.Status_STATUS_RECOGNITION_FAILED_PROCESSING), "Label": "Recognition Failed (Processing)"},
			{"Value": int32(proto.Status_STATUS_RECOGNITION_FAILED_TIMEOUT), "Label": "Recognition Failed (Timeout)"},
		},
	})
}

func UpdateRecognitionTask(c *gin.Context) {
	id := c.Param("id")
	var task proto.DataRecognitionTaskORM
	if err := db.DB.First(&task, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.HTML(http.StatusInternalServerError, "recognition_task/edit.html", gin.H{
			"Error": "Failed to fetch task",
		})
		return
	}

	status := c.PostForm("status")
	if status != "" {
		task.Status = proto.Status_value[status]
	}

	if err := db.DB.Save(&task).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "recognition_task/edit.html", gin.H{
			"Error": "Failed to update task",
			"Task":  task,
		})
		return
	}

	c.Redirect(http.StatusFound, "/recognition-tasks")
}
