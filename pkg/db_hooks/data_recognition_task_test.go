package db_hooks

import (
	"github.com/aws/smithy-go/ptr"
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"github.com/bazilio91/sferra-cloud/pkg/testutils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func createTestTask(db *gorm.DB, status proto.Status, quota int64, sourceImages []string, processedImages []string) (*proto.DataRecognitionTaskORM, error) {
	client, err := testutils.CreateTestClient(DB, "Test Client", quota)
	if err != nil {
		return nil, err
	}

	// Convert client to ORM
	clientORM, err := client.ToORM(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Create task ORM directly
	taskORM := &proto.DataRecognitionTaskORM{
		Id:              uuid.New().String(),
		Client:          &clientORM,
		ClientId:        ptr.Uint64(client.Id),
		Status:          int32(status),
		SourceImages:    sourceImages,
		ProcessedImages: processedImages,
		CreatedAt:       &now,
		UpdatedAt:       &now,
	}

	return taskORM, nil
}

var _ = Describe("DataRecognitionTask", func() {
	BeforeEach(func() {
		testutils.ClearDatabase(DB)
	})

	Describe("ready for processing state", func() {

		It("should fail if quota is below 1", func() {
			task, err := createTestTask(DB, proto.Status_STATUS_READY_FOR_PROCESSING, 0, []string{"image1.jpg"}, nil)
			Expect(err).NotTo(HaveOccurred())

			err = DB.Save(task).Error
			Expect(err).NotTo(HaveOccurred())

			// Reload task to get updated state
			err = DB.Model(&proto.DataRecognitionTaskORM{}).Preload("Client").First(task, "id = ?", task.Id).Error
			Expect(err).NotTo(HaveOccurred())

			Expect(task.Status).To(Equal(int32(proto.Status_STATUS_IMAGES_FAILED_QUOTA)))
			Expect(task.Error).To(Equal("insufficient quota"))
		})

		It("should move to next state if quota is low but non-zero", func() {
			task, err := createTestTask(DB, proto.Status_STATUS_READY_FOR_PROCESSING, 1, []string{"image1.jpg", "image2.jpg"}, nil)
			Expect(err).NotTo(HaveOccurred())

			err = DB.Save(task).Error
			Expect(err).NotTo(HaveOccurred())

			// Reload task to get updated state
			err = DB.Model(&proto.DataRecognitionTaskORM{}).Preload("Client").First(task, "id = ?", task.Id).Error
			Expect(err).NotTo(HaveOccurred())

			Expect(task.Status).To(Equal(int32(proto.Status_STATUS_IMAGES_PENDING)))
		})

		It("should move to next state if quota is ok", func() {
			task, err := createTestTask(DB, proto.Status_STATUS_READY_FOR_PROCESSING, 100, []string{"image1.jpg", "image2.jpg"}, nil)
			Expect(err).NotTo(HaveOccurred())

			err = DB.Save(task).Error
			Expect(err).NotTo(HaveOccurred())

			// Reload task to get updated state
			err = DB.Model(&proto.DataRecognitionTaskORM{}).Preload("Client").First(task, "id = ?", task.Id).Error
			Expect(err).NotTo(HaveOccurred())

			Expect(task.Status).To(Equal(int32(proto.Status_STATUS_IMAGES_PENDING)))
			// check if quota is deducted
			var client proto.ClientORM
			err = DB.Model(&proto.ClientORM{}).First(&client, task.Client.Id).Error
			Expect(err).NotTo(HaveOccurred())
			Expect(client.Quota).To(Equal(int64(98)))
		})

		It("should set in error state if no source images provided", func() {
			task, err := createTestTask(DB, proto.Status_STATUS_READY_FOR_PROCESSING, 100, nil, nil)
			Expect(err).NotTo(HaveOccurred())

			err = DB.Save(task).Error
			Expect(err).NotTo(HaveOccurred())

			// Reload task to get updated state
			err = DB.Model(&proto.DataRecognitionTaskORM{}).Preload("Client").First(task, "id = ?", task.Id).Error
			Expect(err).NotTo(HaveOccurred())

			Expect(task.Status).To(Equal(int32(proto.Status_STATUS_IMAGES_FAILED_PROCESSING)))
			Expect(task.Error).To(Equal("no images provided"))
		})
	})

	Describe("created state", func() {
		It("should not change state", func() {
			task, err := createTestTask(DB, proto.Status_STATUS_CREATED, 100, []string{"image1.jpg"}, nil)
			Expect(err).NotTo(HaveOccurred())

			err = DB.Save(task).Error
			Expect(err).NotTo(HaveOccurred())

			// Reload task to get updated state
			err = DB.Model(&proto.DataRecognitionTaskORM{}).Preload("Client").First(task, "id = ?", task.Id).Error
			Expect(err).NotTo(HaveOccurred())

			// Status should remain STATUS_CREATED as we wait for user to set it to STATUS_READY_FOR_PROCESSING
			Expect(task.Status).To(Equal(int32(proto.Status_STATUS_CREATED)))
			Expect(task.Error).To(BeEmpty())
		})
	})
})
