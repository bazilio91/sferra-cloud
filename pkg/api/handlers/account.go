package handlers

import (
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"net/http"

	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/gin-gonic/gin"
)

// @Summary GetTaskImage Account Info
// @Description GetTaskImage information about the current user
// @Tags account
// @Produce json
// @Success 200 {object} AccountInfoResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/account [get]
func GetAccountInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	var user proto.ClientUser
	if err := db.DB.Preload("Client").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Could not fetch user"})
		return
	}

	c.JSON(http.StatusOK, AccountInfoResponse{
		User: user,
	})
}

type AccountInfoResponse struct {
	User proto.ClientUser `json:"user"`
}
