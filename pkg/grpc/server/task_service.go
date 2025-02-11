package server

import (
	"context"
	"github.com/bazilio91/sferra-cloud/pkg/db_hooks"
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"github.com/bazilio91/sferra-cloud/pkg/types"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TaskService struct {
	proto.UnimplementedTaskServiceServer
	db           *gorm.DB
	stateMachine *db_hooks.StateMachine
}

func NewTaskService(db *gorm.DB, machine *db_hooks.StateMachine) *TaskService {
	return &TaskService{
		db:           db,
		stateMachine: machine,
	}
}

func (s *TaskService) Subscribe(req *proto.SubscribeRequest, stream proto.TaskService_SubscribeServer) error {
	subscriberId := uuid.New().String()
	taskChan := s.stateMachine.Subscribe(subscriberId, req.Queue)
	defer s.stateMachine.Unsubscribe(subscriberId)

	// Query existing pending tasks
	var existingTasks []proto.DataRecognitionTaskORM
	subscribeToStatus := proto.Status_STATUS_IMAGES_PENDING
	if req.Queue == proto.Queues_QUEUE_DATA_RECOGNITION {
		subscribeToStatus = proto.Status_STATUS_RECOGNITION_PENDING
	}

	if err := s.db.Where("status = ?", int32(subscribeToStatus)).Find(&existingTasks).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to query existing tasks")
	}

	// Send existing tasks first
	for _, task := range existingTasks {
		if err := stream.Send(&proto.SubscribeTaskResponse{
			TaskId: task.Id,
		}); err != nil {
			return status.Errorf(codes.Internal, "failed to send task")
		}
	}

	// Listen for new tasks
	for {
		select {
		case task, ok := <-taskChan:
			if !ok {
				// Channel was closed
				return nil
			}
			if err := stream.Send(&proto.SubscribeTaskResponse{TaskId: task.Id}); err != nil {
				return status.Errorf(codes.Internal, "failed to send task")
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (s *TaskService) ReserveTask(ctx context.Context, req *proto.ReserveTaskRequest) (*proto.ReserveTaskResponse, error) {
	var task proto.DataRecognitionTaskORM
	if err := s.db.First(&task, "id = ?", req.TaskId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &proto.ReserveTaskResponse{Success: false}, status.Errorf(codes.NotFound, "task not found")
		}
		return &proto.ReserveTaskResponse{Success: false}, status.Errorf(codes.Internal, "database error")
	}

	// Only allow reserving tasks in pending states
	if types.IsTerminalState(proto.Status(task.Status)) {
		return &proto.ReserveTaskResponse{Success: false}, status.Errorf(codes.FailedPrecondition, "task is in terminal state")
	}

	task.WorkerId = req.WorkerId
	if err := s.db.Save(&task).Error; err != nil {
		return &proto.ReserveTaskResponse{Success: false}, status.Errorf(codes.Internal, "failed to update task")
	}

	return &proto.ReserveTaskResponse{Success: true}, nil
}

func (s *TaskService) ReportTaskStatus(ctx context.Context, req *proto.ReportTaskStatusRequest) (*proto.Ack, error) {
	var task proto.DataRecognitionTaskORM
	if err := s.db.First(&task, "id = ?", req.TaskId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &proto.Ack{Success: false}, status.Errorf(codes.NotFound, "task not found")
		}
		return &proto.Ack{Success: false}, status.Errorf(codes.Internal, "database error")
	}

	task.StatusText = req.Status
	if err := s.db.Save(&task).Error; err != nil {
		return &proto.Ack{Success: false}, status.Errorf(codes.Internal, "failed to update task")
	}

	// Process state machine
	if err := s.stateMachine.Process(ctx, &task); err != nil {
		return &proto.Ack{Success: false}, status.Errorf(codes.Internal, "failed to process task state")
	}

	return &proto.Ack{Success: true}, nil
}

func (s *TaskService) FinishTask(ctx context.Context, req *proto.FinishTaskRequest) (*proto.Ack, error) {
	var taskOrm proto.DataRecognitionTaskORM
	if err := s.db.First(&taskOrm, "id = ?", req.Id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &proto.Ack{Success: false}, status.Errorf(codes.NotFound, "task not found")
		}
		return &proto.Ack{Success: false}, status.Errorf(codes.Internal, "database error")
	}

	switch proto.Status(taskOrm.Status) {
	case proto.Status_STATUS_IMAGES_PROCESSING:
		taskOrm.Status = int32(proto.Status_STATUS_IMAGES_COMPLETED)
		taskOrm.ProcessedImages = req.ProcessedImages
	case proto.Status_STATUS_RECOGNITION_PROCESSING:
		taskOrm.Status = int32(proto.Status_STATUS_RECOGNITION_COMPLETED)
		result := datatypes.NewJSONType[proto.TreeNode](*req.RecognitionResult)
		taskOrm.RecognitionResult = &result
	}

	if err := s.db.Save(&taskOrm).Error; err != nil {
		return &proto.Ack{Success: false}, status.Errorf(codes.Internal, "failed to update task")
	}

	// Process state machine
	if err := s.stateMachine.Process(ctx, &taskOrm); err != nil {
		return &proto.Ack{Success: false}, status.Errorf(codes.Internal, "failed to process task state")
	}

	return &proto.Ack{Success: true}, nil
}

func (s *TaskService) FailTask(ctx context.Context, task *proto.DataRecognitionTask) (*proto.Ack, error) {
	var taskOrm proto.DataRecognitionTaskORM
	if err := s.db.First(&taskOrm, "id = ?", task.Id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &proto.Ack{Success: false}, status.Errorf(codes.NotFound, "task not found")
		}
		return &proto.Ack{Success: false}, status.Errorf(codes.Internal, "database error")
	}

	switch proto.Status(taskOrm.Status) {
	case proto.Status_STATUS_IMAGES_PROCESSING:
		taskOrm.Status = int32(proto.Status_STATUS_IMAGES_FAILED_PROCESSING)
	case proto.Status_STATUS_RECOGNITION_PROCESSING:
		taskOrm.Status = int32(proto.Status_STATUS_RECOGNITION_FAILED_PROCESSING)
	}

	taskOrm.Error = task.Error
	if err := s.db.Save(&taskOrm).Error; err != nil {
		return &proto.Ack{Success: false}, status.Errorf(codes.Internal, "failed to update task")
	}

	// Process state machine
	if err := s.stateMachine.Process(ctx, &taskOrm); err != nil {
		return &proto.Ack{Success: false}, status.Errorf(codes.Internal, "failed to process task state")
	}

	return &proto.Ack{Success: true}, nil
}
