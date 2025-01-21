package db_hooks

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"github.com/bazilio91/sferra-cloud/pkg/types"
)

const (
	ImageProcessingTimeout = 10 * time.Minute
	RecognitionTimeout     = 15 * time.Minute
)

var (
	ErrInvalidTransition = errors.New("invalid state transition")
	ErrInsufficientQuota = errors.New("insufficient quota")
	ErrQuotaExceeded     = errors.New("quota exceeded")
	ErrNoImages          = errors.New("no images provided")
)

type TaskSubscriber struct {
	Queue    proto.Queues
	TaskChan chan *proto.DataRecognitionTaskORM
}

// StateMachine handles the state transitions for DataRecognitionTask
type StateMachine struct {
	db          *gorm.DB
	subscribers map[string]*TaskSubscriber
	mu          sync.RWMutex
}

// NewStateMachine creates a new state machine instance
func NewStateMachine(db *gorm.DB) *StateMachine {
	sm := &StateMachine{
		db:          db,
		subscribers: make(map[string]*TaskSubscriber),
	}

	sm.registerDataRecognitionTaskHooks()

	return sm
}

// Subscribe adds a new subscriber for task updates
func (sm *StateMachine) Subscribe(subscriberId string, queue proto.Queues) chan *proto.DataRecognitionTaskORM {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	taskChan := make(chan *proto.DataRecognitionTaskORM, 100)
	sm.subscribers[subscriberId] = &TaskSubscriber{
		Queue:    queue,
		TaskChan: taskChan,
	}

	return taskChan
}

// Unsubscribe removes a subscriber
func (sm *StateMachine) Unsubscribe(subscriberId string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sub, exists := sm.subscribers[subscriberId]; exists {
		close(sub.TaskChan)
		delete(sm.subscribers, subscriberId)
	}
}

// notifySubscribers sends task updates to all relevant subscribers
func (sm *StateMachine) notifySubscribers(task *proto.DataRecognitionTaskORM) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, sub := range sm.subscribers {
		// Filter by queue
		if (sub.Queue == proto.Queues_QUEUE_IMAGE_PROCESSING && proto.Status(task.Status) == proto.Status_STATUS_IMAGES_PENDING) ||
			(sub.Queue == proto.Queues_QUEUE_DATA_RECOGNITION && proto.Status(task.Status) == proto.Status_STATUS_RECOGNITION_PENDING) {
			// Non-blocking send
			select {
			case sub.TaskChan <- task:
			default:
				// Channel is full, skip this notification
			}
		}
	}
}

// Process analyzes the current state of the task and performs the next appropriate action
func (sm *StateMachine) Process(ctx context.Context, task *proto.DataRecognitionTaskORM) error {
	// Skip processing for terminal states
	if types.IsTerminalState(proto.Status(task.Status)) {
		return nil
	}

	// Load client if not loaded
	if task.Client == nil {
		var client proto.ClientORM
		if err := sm.db.First(&client, task.ClientId).Error; err != nil {
			return fmt.Errorf("failed to load client: %w", err)
		}
		task.Client = &client
	}

	switch proto.Status(task.Status) {
	case proto.Status_STATUS_CREATED:
		return sm.handleCreatedState(ctx, task)
	case proto.Status_STATUS_READY_FOR_PROCESSING:
		return sm.handleReadyForProcessing(ctx, task)
	case proto.Status_STATUS_IMAGES_PENDING:
		return sm.handleImagesPending(ctx, task)
	case proto.Status_STATUS_IMAGES_PROCESSING:
		return sm.handleImagesProcessing(ctx, task)
	case proto.Status_STATUS_IMAGES_COMPLETED:
		return sm.handleImagesCompleted(ctx, task)
	case proto.Status_STATUS_RECOGNITION_PENDING:
		return sm.handleRecognitionPending(ctx, task)
	case proto.Status_STATUS_RECOGNITION_PROCESSING:
		return sm.handleRecognitionProcessing(ctx, task)
	case proto.Status_STATUS_RECOGNITION_FAILED_QUOTA:
		return sm.handleRecognitionFailedQuota(ctx, task)
	case proto.Status_STATUS_IMAGES_FAILED_QUOTA:
		return sm.handleImagesFailedQuota(ctx, task)
	default:
		// Terminal states
		return nil
	}
}

func (sm *StateMachine) handleCreatedState(ctx context.Context, task *proto.DataRecognitionTaskORM) error {
	// In CREATED state, we just wait for the user to set STATUS_READY_FOR_PROCESSING via API
	// No automatic transitions from this state
	return nil
}

func (sm *StateMachine) handleReadyForProcessing(ctx context.Context, task *proto.DataRecognitionTaskORM) error {
	// Check if client has enough quota
	var client proto.ClientORM

	err := sm.db.Model(&task).Association("Client").Find(&client)

	if err != nil {
		return fmt.Errorf("failed to get client: %s", err)
	}

	if len(task.SourceImages) == 0 {
		task.Error = "no images provided"
		task.Status = int32(proto.Status_STATUS_IMAGES_FAILED_PROCESSING)
		return sm.db.Save(task).Error
	}

	// Check quota
	if task.Client.Quota <= 0 {
		task.Error = "insufficient quota"
		task.Status = int32(proto.Status_STATUS_IMAGES_FAILED_QUOTA)
		return sm.db.Save(task).Error
	}

	// Deduct quota
	task.Client.Quota -= int64(len(task.SourceImages))
	if err := sm.db.Save(&task.Client).Error; err != nil {
		return fmt.Errorf("failed to update client quota: %w", err)
	}

	// Move to images pending state
	task.Status = int32(proto.Status_STATUS_IMAGES_PENDING)
	return sm.db.Save(task).Error
}

func (sm *StateMachine) handleImagesPending(ctx context.Context, task *proto.DataRecognitionTaskORM) error {
	sm.notifySubscribers(task)
	return nil
}

func (sm *StateMachine) handleImagesProcessing(ctx context.Context, task *proto.DataRecognitionTaskORM) error {
	// Check for timeout
	if time.Since(*task.UpdatedAt) > ImageProcessingTimeout {
		task.Error = "timeout"
		task.Status = int32(proto.Status_STATUS_IMAGES_FAILED_TIMEOUT)
		return sm.db.Save(task).Error
	}

	// Check if all images are processed
	if len(task.ProcessedImages) == len(task.SourceImages) {
		task.Status = int32(proto.Status_STATUS_IMAGES_COMPLETED)
		return sm.db.Save(task).Error
	}

	return nil
}

func (sm *StateMachine) handleImagesCompleted(ctx context.Context, task *proto.DataRecognitionTaskORM) error {
	// Check if client has enough quota for recognition
	var client proto.ClientORM
	if err := sm.db.First(&client, task.ClientId).Error; err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	// Check quota
	if client.Quota < 1 {
		task.Error = "insufficient quota"
		task.Status = int32(proto.Status_STATUS_RECOGNITION_FAILED_QUOTA)
		return sm.db.Save(task).Error
	}

	// Deduct quota (1 for recognition)
	client.Quota -= 1
	if err := sm.db.Save(&client).Error; err != nil {
		return fmt.Errorf("failed to update client quota: %w", err)
	}

	// Start recognition
	task.Status = int32(proto.Status_STATUS_RECOGNITION_PENDING)
	sm.notifySubscribers(task)
	return sm.db.Save(task).Error
}

func (sm *StateMachine) handleRecognitionPending(ctx context.Context, task *proto.DataRecognitionTaskORM) error {
	sm.notifySubscribers(task)
	return nil
}

func (sm *StateMachine) handleRecognitionProcessing(ctx context.Context, task *proto.DataRecognitionTaskORM) error {
	// Check for timeout
	if time.Since(*task.UpdatedAt) > RecognitionTimeout {
		task.Error = "timeout"
		task.Status = int32(proto.Status_STATUS_RECOGNITION_FAILED_TIMEOUT)
		return sm.db.Save(task).Error
	}

	// In a real implementation, this would check if recognition is complete
	// For now, we'll just mark it as complete
	task.Status = int32(proto.Status_STATUS_RECOGNITION_COMPLETED)
	return sm.db.Save(task).Error
}

func (sm *StateMachine) handleRecognitionFailedQuota(ctx context.Context, task *proto.DataRecognitionTaskORM) error {
	// Handle recognition failed quota state
	return nil
}

func (sm *StateMachine) handleImagesFailedQuota(ctx context.Context, task *proto.DataRecognitionTaskORM) error {
	// Handle images failed quota state
	return nil
}

// IsTerminalStateOld returns true if the status is a terminal state
func IsTerminalStateOld(status proto.Status) bool {
	switch status {
	case proto.Status_STATUS_IMAGES_FAILED_QUOTA,
		proto.Status_STATUS_IMAGES_FAILED_PROCESSING,
		proto.Status_STATUS_IMAGES_FAILED_TIMEOUT,
		proto.Status_STATUS_RECOGNITION_FAILED_QUOTA,
		proto.Status_STATUS_RECOGNITION_FAILED_PROCESSING,
		proto.Status_STATUS_RECOGNITION_FAILED_TIMEOUT,
		proto.Status_STATUS_RECOGNITION_COMPLETED:
		return true
	default:
		return false
	}
}
