package db_hooks

import (
	"context"
	"time"

	"github.com/bazilio91/sferra-cloud/pkg/proto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StateMachine Subscription", func() {
	var (
		sm *StateMachine
	)

	BeforeEach(func() {
		sm = NewStateMachine(DB)
	})

	Context("Subscribe and Notify", func() {
		It("should notify image processing subscribers when task becomes pending", func() {
			// Create subscriber
			subscriberId := "test-subscriber-1"
			taskChan := sm.Subscribe(subscriberId, proto.Queues_QUEUE_IMAGE_PROCESSING)
			defer sm.Unsubscribe(subscriberId)

			// Create a task that will transition to IMAGES_PENDING
			task, err := createTestTask(DB, proto.Status_STATUS_CREATED, 100, []string{"test.jpg"}, nil)
			Expect(err).To(BeNil())
			Expect(DB.Create(task).Error).To(BeNil())

			task.Status = int32(proto.Status_STATUS_IMAGES_PENDING)
			err = DB.Save(&task).Error
			Expect(err).To(BeNil())

			err = sm.Process(context.Background(), task)
			Expect(err).To(BeNil())
			Expect(proto.Status(task.Status)).To(Equal(proto.Status_STATUS_IMAGES_PENDING))

			// Check if subscriber received the task
			var receivedTask *proto.DataRecognitionTaskORM
			Eventually(taskChan, 2*time.Second).Should(Receive(&receivedTask))
			Expect(receivedTask.Id).To(Equal(task.Id))
		})

		It("should notify recognition subscribers when task becomes pending", func() {
			// Create subscriber
			subscriberId := "test-subscriber-2"
			taskChan := sm.Subscribe(subscriberId, proto.Queues_QUEUE_DATA_RECOGNITION)
			defer sm.Unsubscribe(subscriberId)

			// Create a task that will transition to RECOGNITION_PENDING
			task, err := createTestTask(DB, proto.Status_STATUS_IMAGES_COMPLETED, 100, []string{"test.jpg"}, []string{"processed.jpg"})
			Expect(err).To(BeNil())
			Expect(DB.Create(task).Error).To(BeNil())

			// Process task to transition to RECOGNITION_PENDING
			err = sm.Process(context.Background(), task)
			Expect(err).To(BeNil())
			Expect(proto.Status(task.Status)).To(Equal(proto.Status_STATUS_RECOGNITION_PENDING))

			// Check if subscriber received the task
			var receivedTask *proto.DataRecognitionTaskORM
			Eventually(taskChan, 2*time.Second).Should(Receive(&receivedTask))
			Expect(receivedTask.Id).To(Equal(task.Id))
		})

		It("should not notify subscribers for non-pending states", func() {
			// Create subscribers for both queues
			subId1 := "test-subscriber-3"
			subId2 := "test-subscriber-4"
			imgChan := sm.Subscribe(subId1, proto.Queues_QUEUE_IMAGE_PROCESSING)
			recChan := sm.Subscribe(subId2, proto.Queues_QUEUE_DATA_RECOGNITION)
			defer sm.Unsubscribe(subId1)
			defer sm.Unsubscribe(subId2)

			// Create a task in processing state
			task, err := createTestTask(DB, proto.Status_STATUS_IMAGES_PROCESSING, 100, []string{"test.jpg"}, nil)
			Expect(err).To(BeNil())
			Expect(DB.Create(task).Error).To(BeNil())

			// Process task
			err = sm.Process(context.Background(), task)
			Expect(err).To(BeNil())

			// Verify no notifications were sent
			Consistently(imgChan, 1*time.Second).ShouldNot(Receive())
			Consistently(recChan, 1*time.Second).ShouldNot(Receive())
		})

		It("should handle multiple subscribers for the same queue", func() {
			// Create multiple subscribers
			subId1 := "test-subscriber-5"
			subId2 := "test-subscriber-6"
			chan1 := sm.Subscribe(subId1, proto.Queues_QUEUE_IMAGE_PROCESSING)
			chan2 := sm.Subscribe(subId2, proto.Queues_QUEUE_IMAGE_PROCESSING)
			defer sm.Unsubscribe(subId1)
			defer sm.Unsubscribe(subId2)

			// Create and process task
			task, err := createTestTask(DB, proto.Status_STATUS_CREATED, 100, []string{"test.jpg"}, nil)
			Expect(err).To(BeNil())
			Expect(DB.Create(task).Error).To(BeNil())

			task.Status = int32(proto.Status_STATUS_IMAGES_PENDING)
			Expect(DB.Save(task).Error).To(BeNil())

			// Both subscribers should receive the task
			var receivedTask1, receivedTask2 *proto.DataRecognitionTaskORM
			Eventually(chan1, 2*time.Second).Should(Receive(&receivedTask1))
			Eventually(chan2, 2*time.Second).Should(Receive(&receivedTask2))
			Expect(receivedTask1.Id).To(Equal(task.Id))
			Expect(receivedTask2.Id).To(Equal(task.Id))
		})

		It("should handle unsubscribe correctly", func() {
			// Create subscriber
			subscriberId := "test-subscriber-7"
			taskChan := sm.Subscribe(subscriberId, proto.Queues_QUEUE_IMAGE_PROCESSING)

			// Unsubscribe
			sm.Unsubscribe(subscriberId)

			// Create and process task
			task, err := createTestTask(DB, proto.Status_STATUS_CREATED, 100, []string{"test.jpg"}, nil)
			Expect(err).To(BeNil())
			Expect(DB.Create(task).Error).To(BeNil())

			err = sm.Process(context.Background(), task)
			Expect(err).To(BeNil())

			// Channel should be closed
			_, ok := <-taskChan
			Expect(ok).To(BeFalse())
		})
	})
})
