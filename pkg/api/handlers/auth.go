package handlers

import (
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// @Summary Login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param login body LoginInput true "Login credentials"
// @Success 200 {object} TokenResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	var user proto.ClientUser
	if err := db.DB.Preload("Client").Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := jwtManager.GenerateToken(user.Id, user.ClientId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{Token: token})
}

// @Summary Register
// @Description Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param register body RegisterInput true "Registration data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/register [post]
func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Check if client exists
	var client proto.Client
	if err := db.DB.First(&client, input.ClientID).Error; err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid Client ID"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Could not create user"})
		return
	}

	user := proto.ClientUser{
		Email:    input.Email,
		Password: string(hashedPassword),
		ClientId: client.Id,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Could not create user"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "User created successfully"})
}

// Input and Output structs
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	ClientID uint64 `json:"clientID" binding:"required"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
