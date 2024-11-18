package handlers

import (
	"github.com/bazilio91/sferra-cloud/pkg/models"
	"net/http"

	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/gin-gonic/gin"
)

// @Summary Get Account Info
// @Description Get information about the current user
// @Tags account
// @Produce json
// @Success 200 {object} AccountInfoResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /account [get]
func GetAccountInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Could not fetch user"})
		return
	}

	c.JSON(http.StatusOK, AccountInfoResponse{
		Email: user.Email,
		// Include other account details as needed
	})
}

type AccountInfoResponse struct {
	Email string `json:"email"`
}
