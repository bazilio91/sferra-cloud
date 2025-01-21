package handlers_test

import (
	"bytes"
	"encoding/json"
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"net/http"
	"net/http/httptest"

	"github.com/bazilio91/sferra-cloud/pkg/api/handlers"
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/bazilio91/sferra-cloud/pkg/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth Handlers", func() {
	var (
		client *proto.Client
	)

	BeforeEach(func() {
		testutils.ClearDatabase(DB)

		var err error
		client, err = testutils.CreateTestClient(DB, "Test Client", 100)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Login", func() {
		Context("With valid credentials", func() {
			It("should return a token", func() {
				// Create test user
				_, err := testutils.CreateTestUser(DB, "test@example.com", "password123", client.Id)
				Expect(err).NotTo(HaveOccurred())

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
				Expect(resp.Code).To(Equal(http.StatusOK))
				var response handlers.TokenResponse
				json.Unmarshal(resp.Body.Bytes(), &response)
				Expect(response.Token).NotTo(BeEmpty())
			})
		})

		Context("With invalid credentials", func() {
			It("should return an error", func() {
				// Create test user with a different password
				_, err := testutils.CreateTestUser(DB, "test@example.com", "password123", client.Id)
				Expect(err).NotTo(HaveOccurred())

				// Create test request with incorrect password
				loginInput := handlers.LoginInput{
					Email:    "test@example.com",
					Password: "wrongpassword",
				}
				jsonValue, _ := json.Marshal(loginInput)
				req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonValue))
				req.Header.Set("Content-Type", "application/json")

				// Perform request
				resp := httptest.NewRecorder()
				r.ServeHTTP(resp, req)

				// Assertions
				Expect(resp.Code).To(Equal(http.StatusUnauthorized))
				var response handlers.ErrorResponse
				json.Unmarshal(resp.Body.Bytes(), &response)
				Expect(response.Error).To(Equal("Invalid email or password"))
			})
		})
	})

	Describe("Register", func() {
		Context("With valid input", func() {
			It("should create a new user", func() {
				// Create test request
				registerInput := handlers.RegisterInput{
					Email:    "newuser@example.com",
					Password: "password123",
					ClientID: client.Id,
				}
				jsonValue, _ := json.Marshal(registerInput)
				req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonValue))
				req.Header.Set("Content-Type", "application/json")

				// Perform request
				resp := httptest.NewRecorder()
				r.ServeHTTP(resp, req)

				// Assertions
				Expect(resp.Code).To(Equal(http.StatusOK))
				var response handlers.SuccessResponse
				json.Unmarshal(resp.Body.Bytes(), &response)
				Expect(response.Message).To(Equal("User created successfully"))

				// Verify that the user was created
				var count int64
				err := db.DB.Model(&proto.ClientUser{}).Where("email = ?", "newuser@example.com").Count(&count).Error
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(int64(1)))
			})
		})

		Context("With invalid input", func() {
			It("should return an error when client ID is invalid", func() {
				// Create test request with invalid client ID
				registerInput := handlers.RegisterInput{
					Email:    "newuser@example.com",
					Password: "password123",
					ClientID: 9999, // Non-existent client ID
				}
				jsonValue, _ := json.Marshal(registerInput)
				req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonValue))
				req.Header.Set("Content-Type", "application/json")

				// Perform request
				resp := httptest.NewRecorder()
				r.ServeHTTP(resp, req)

				// Assertions
				Expect(resp.Code).To(Equal(http.StatusBadRequest))
				var response handlers.ErrorResponse
				json.Unmarshal(resp.Body.Bytes(), &response)
				Expect(response.Error).To(Equal("Invalid Client ID"))
			})
		})
	})
})
