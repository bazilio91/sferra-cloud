package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/bazilio91/sferra-cloud/pkg/api/handlers"
	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"github.com/bazilio91/sferra-cloud/pkg/testutils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DataRecognitionTask", func() {
	var (
		claims *auth.Claims
		client proto.Client
	)

	BeforeEach(func() {
		testutils.ClearDatabase(DB)
		claims = &auth.Claims{
			ClientID: 1,
		}
		client = proto.Client{
			Id: uint64(claims.ClientID),
		}

		// Create client in database
		clientORM, err := client.ToORM(context.Background())
		Expect(err).NotTo(HaveOccurred())
		err = db.DB.Create(&clientORM).Error
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("CreateDataRecognitionTask", func() {
		It("should create a new data recognition request", func() {
			request := proto.DataRecognitionTask{
				Client:       &client,
				SourceImages: []string{"image1.jpg", "image2.jpg"},
			}

			jsonData, err := json.Marshal(request)
			Expect(err).NotTo(HaveOccurred())

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/recognition_requests", bytes.NewBuffer(jsonData))
			c.Set("user", claims)

			handlers.CreateDataRecognitionTask(c)

			Expect(w.Code).To(Equal(http.StatusCreated))

			var response proto.DataRecognitionTask
			err = json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Client.Id).To(Equal(uint64(claims.ClientID)))
			Expect(response.Status).To(Equal(proto.Status_STATUS_IMAGES_PENDING))
			Expect(response.SourceImages).To(Equal(request.SourceImages))
		})
	})

	Describe("GetDataRecognitionTask", func() {
		var request proto.DataRecognitionTask

		BeforeEach(func() {
			id := uuid.New().String()
			request = proto.DataRecognitionTask{
				Id:           id,
				Client:       &client,
				SourceImages: []string{"image1.jpg"},
				Status:       proto.Status_STATUS_CREATED,
			}

			// Convert to ORM and save
			ormObj, err := request.ToORM(context.Background())
			Expect(err).NotTo(HaveOccurred())
			err = db.DB.Create(&ormObj).Error
			Expect(err).NotTo(HaveOccurred())
		})

		It("should get an existing data recognition request", func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/recognition_requests/%s", request.Id), nil)
			c.Set("user", claims)
			c.AddParam("id", request.Id)

			handlers.GetDataRecognitionTask(c)

			Expect(w.Code).To(Equal(http.StatusOK))

			var response proto.DataRecognitionTask
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Id).To(Equal(request.Id))
			Expect(response.Client.Id).To(Equal(uint64(claims.ClientID)))
			Expect(response.Status).To(Equal(request.Status))
			Expect(response.SourceImages).To(Equal(request.SourceImages))
		})

		It("should return 404 for non-existent request", func() {
			nonExistentID := uuid.New().String()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/recognition_requests/%s", nonExistentID), nil)
			c.Set("user", claims)
			c.AddParam("id", nonExistentID)

			handlers.GetDataRecognitionTask(c)

			Expect(w.Code).To(Equal(http.StatusNotFound))
		})

		It("should return 400 for invalid UUID format", func() {
			invalidID := "not-a-uuid"
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/recognition_requests/%s", invalidID), nil)
			c.Set("user", claims)
			c.AddParam("id", invalidID)

			handlers.GetDataRecognitionTask(c)

			Expect(w.Code).To(Equal(http.StatusBadRequest))
			var response handlers.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Error).To(ContainSubstring("invalid task ID format"))
		})

		It("should return 403 for request belonging to another client", func() {
			otherClaims := &auth.Claims{
				ClientID: 2,
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/recognition_requests/%s", request.Id), nil)
			c.Set("user", otherClaims)
			c.AddParam("id", request.Id)

			handlers.GetDataRecognitionTask(c)

			Expect(w.Code).To(Equal(http.StatusForbidden))
		})
	})

	Describe("ListDataRecognitionTask", func() {
		BeforeEach(func() {
			// Create multiple requests
			for i := 1; i <= 3; i++ {
				id := uuid.New().String()
				request := proto.DataRecognitionTask{
					Id:           id,
					Client:       &client,
					SourceImages: []string{fmt.Sprintf("image%d.jpg", i)},
					Status:       proto.Status_STATUS_CREATED,
				}

				ormObj, err := request.ToORM(context.Background())
				Expect(err).NotTo(HaveOccurred())
				err = db.DB.Create(&ormObj).Error
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should list all data recognition requests for the client", func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/recognition_requests", nil)
			c.Set("claims", claims)

			handlers.ListDataRecognitionTask(c)

			Expect(w.Code).To(Equal(http.StatusOK))

			var response handlers.DataRecognitionTaskListResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(response.Results)).To(Equal(3))
			for _, req := range response.Results {
				Expect(req.Client.Id).To(Equal(uint64(claims.ClientID)))
			}
		})
	})

	Describe("DeleteDataRecognitionTask", func() {
		var request proto.DataRecognitionTask

		BeforeEach(func() {
			id := uuid.New().String()
			request = proto.DataRecognitionTask{
				Id:           id,
				Client:       &client,
				SourceImages: []string{"image1.jpg"},
				Status:       proto.Status_STATUS_CREATED,
			}

			// Convert to ORM and save
			ormObj, err := request.ToORM(context.Background())
			Expect(err).NotTo(HaveOccurred())
			err = db.DB.Create(&ormObj).Error
			Expect(err).NotTo(HaveOccurred())
		})

		It("should delete an existing data recognition request", func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/recognition_requests/%s", request.Id), nil)
			c.Set("user", claims)
			c.AddParam("id", request.Id)

			handlers.DeleteDataRecognitionTask(c)

			Expect(w.Code).To(Equal(http.StatusOK))

			// Verify deletion
			var count int64
			err := db.DB.Model(&proto.DataRecognitionTaskORM{}).Where("id = ?", request.Id).Count(&count).Error
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(int64(0)))
		})

		It("should return 404 for non-existent request", func() {
			nonExistentID := uuid.New().String()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/recognition_requests/%s", nonExistentID), nil)
			c.Set("user", claims)
			c.AddParam("id", nonExistentID)

			handlers.DeleteDataRecognitionTask(c)

			Expect(w.Code).To(Equal(http.StatusNotFound))
		})

		It("should return 400 for invalid UUID format", func() {
			invalidID := "not-a-uuid"
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/recognition_requests/%s", invalidID), nil)
			c.Set("user", claims)
			c.AddParam("id", invalidID)

			handlers.DeleteDataRecognitionTask(c)

			Expect(w.Code).To(Equal(http.StatusBadRequest))
			var response handlers.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Error).To(ContainSubstring("invalid task ID format"))
		})

		It("should return 403 for request belonging to another client", func() {
			otherClaims := &auth.Claims{
				ClientID: 2,
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/recognition_requests/%s", request.Id), nil)
			c.Set("user", otherClaims)
			c.AddParam("id", request.Id)

			handlers.DeleteDataRecognitionTask(c)

			Expect(w.Code).To(Equal(http.StatusForbidden))
		})
	})
})
