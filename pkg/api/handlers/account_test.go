package handlers_test

import (
	"context"
	"encoding/json"
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

func TestGetAccountInfo(t *testing.T) {
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

	// Generate token
	token, err := jwtManager.GenerateJWT(1) // Assuming the user ID is 1
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create test request
	req, _ := http.NewRequest("GET", "/api/v1/account", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Perform request
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// Assertions
	assert.Equal(t, http.StatusOK, resp.Code)
	var response handlers.AccountInfoResponse
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	assert.Equal(t, "test@example.com", response.Email)
}
