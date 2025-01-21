package handlers_test

import (
	"encoding/json"
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"net/http"
	"net/http/httptest"

	"github.com/bazilio91/sferra-cloud/pkg/api/handlers"
	"github.com/bazilio91/sferra-cloud/pkg/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Account Handlers", func() {
	var (
		clientModel *proto.Client
		userModel   *proto.ClientUser
	)

	BeforeEach(func() {
		testutils.ClearDatabase(DB)

		var err error
		clientModel, err = testutils.CreateTestClient(DB, "Test Client", 100)
		Expect(err).NotTo(HaveOccurred())

		userModel, err = testutils.CreateTestUser(DB, "test@example.com", "password123", clientModel.Id)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("GetAccountInfo", func() {
		Context("With valid token", func() {
			It("should return account information", func() {
				// Generate token
				token, err := jwtManager.GenerateToken(userModel.Id, userModel.ClientId)
				Expect(err).NotTo(HaveOccurred())

				// Create test request
				req, _ := http.NewRequest("GET", "/api/v1/account", nil)
				req.Header.Set("Authorization", "Bearer "+token)

				// Perform request
				resp := httptest.NewRecorder()
				r.ServeHTTP(resp, req)

				// Assertions
				Expect(resp.Code).To(Equal(http.StatusOK))
				var response handlers.AccountInfoResponse
				err = json.Unmarshal(resp.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.User.Email).To(Equal("test@example.com"))
				Expect(response.User.Client.Quota).To(Equal(clientModel.Quota))
			})
		})

		Context("With invalid token", func() {
			It("should return unauthorized error", func() {
				// Create test request without token
				req, _ := http.NewRequest("GET", "/api/v1/account", nil)

				// Perform request
				resp := httptest.NewRecorder()
				r.ServeHTTP(resp, req)

				// Assertions
				Expect(resp.Code).To(Equal(http.StatusUnauthorized))
				var response handlers.ErrorResponse
				json.Unmarshal(resp.Body.Bytes(), &response)
				Expect(response.Error).To(Equal("Authorization header missing"))
			})
		})
	})
})
