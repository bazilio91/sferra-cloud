package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/bazilio91/sferra-cloud/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bazilio91/sferra-cloud/pkg/api/handlers"
	"github.com/bazilio91/sferra-cloud/pkg/api/router"
	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/bazilio91/sferra-cloud/pkg/testutils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	testutils.SetupTest(m)
}

func TestLogin(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	ctx := context.Background()

	// Start test database container
	testDB, err := testutils.StartTestDBContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test database: %v", err)
	}
	defer testutils.StopTestDBContainer(ctx, testDB)

	// Clear database before test
	testutils.ClearDatabase()

	// Create test user
	err = testutils.CreateTestUser("test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(testDB.Config.JWTSecret)
	handlers.SetJWTManager(jwtManager)
	handlers.SetConfig(testDB.Config)

	// Create router
	r := router.SetupRouter(jwtManager, testDB.Config)

	// Create test request
	loginInput := handlers.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonValue, _ := json.Marshal(loginInput)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// Assertions
	assert.Equal(t, http.StatusOK, resp.Code)
	var response handlers.TokenResponse
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NotEmpty(t, response.Token)
}

func TestRegister(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	ctx := context.Background()

	// Start test database container
	testDB, err := testutils.StartTestDBContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test database: %v", err)
	}
	defer testutils.StopTestDBContainer(ctx, testDB)

	// Clear database before test
	testutils.ClearDatabase()

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(testDB.Config.JWTSecret)
	handlers.SetJWTManager(jwtManager)
	handlers.SetConfig(testDB.Config)

	// Create router
	r := router.SetupRouter(jwtManager, testDB.Config)

	// Create test request
	registerInput := handlers.RegisterInput{
		Email:    "newuser@example.com",
		Password: "password123",
	}
	jsonValue, _ := json.Marshal(registerInput)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// Assertions
	assert.Equal(t, http.StatusOK, resp.Code)
	var response handlers.SuccessResponse
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "User created successfully", response.Message)

	// Verify that the user was created
	var count int64
	//db := testDB.Config
	//db.DBHost = testDB.Config.DBHost
	//db.DBPort = testDB.Config.DBPort

	// Initialize the database connection
	dsn := db.GetDSN(testDB.Config)
	dbInstance, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Count users with the new email
	dbInstance.Model(&models.User{}).Where("email = ?", "newuser@example.com").Count(&count)
	assert.Equal(t, int64(1), count)
}
