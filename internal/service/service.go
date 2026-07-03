package service

import (
	"context"
	"fmt"
	"go-concurrency-task/internal/domain"
	"go-concurrency-task/internal/dto"
	"go-concurrency-task/internal/repository"
	"go-concurrency-task/internal/workers"

	"github.com/google/uuid"
)

type TaskService struct {
	TaskRepo  *repository.TaskStateRepository
	TaskQueue *workers.TaskManager
	SSEBroker *workers.Broker
}

func NewTaskService(taskRepo *repository.TaskStateRepository, taskQueue *workers.TaskManager, sseBroker *workers.Broker) *TaskService {
	return &TaskService{
		TaskRepo:  taskRepo,
		TaskQueue: taskQueue,
		SSEBroker: sseBroker,
	}
}

func (s *TaskService) CreateTask(req *dto.TaskRequest) (*dto.TaskResponse, error) {
	task := domain.NewTask(uuid.New(), req.Type, req.Payload)
	state := domain.NewTaskState(task)
	if err := s.TaskRepo.Add(state); err != nil {
		return nil, err
	}

	s.TaskQueue.AdToQueue(task.ID)

	res := &dto.TaskResponse{
		ID:     state.Task.ID,
		Status: string(state.Status),
	}

	return res, nil
}

func (s *TaskService) TaskById(ID uuid.UUID) (*dto.TaskResponse, error) {
	state, err := s.TaskRepo.GetTaskState(ID)
	if err != nil {
		return nil, err
	}

	res := &dto.TaskResponse{
		ID:     state.Task.ID,
		Status: string(state.Status),
	}

	return res, nil
}

func (s *TaskService) DeleteTaskById(ID uuid.UUID) (*dto.TaskResponse, error) {
	if err := s.TaskQueue.CancelTask(ID); err != nil {
		return nil, err
	}
	if err := s.TaskRepo.CancelTask(ID); err != nil {
		return nil, err
	}
	s.SSEBroker.WriteMessage(fmt.Sprintf("%v is deleted", ID))

	state, err := s.TaskRepo.GetTaskState(ID)
	if err != nil {
		return nil, err
	}

	res := &dto.TaskResponse{
		ID:     state.Task.ID,
		Status: string(state.Status),
	}

	return res, nil
}

func (s *TaskService) GetAllTasks() ([]*dto.TaskResponse, error) {
	tasks, err := s.TaskRepo.GetTasks()
	if err != nil {
		return nil, err
	}

	var res []*dto.TaskResponse
	for _, taskState := range tasks {
		resTask := &dto.TaskResponse{
			ID:     taskState.Task.ID,
			Status: string(taskState.Status),
		}
		res = append(res, resTask)
	}

	return res, nil
}

func (s *TaskService) GetEvents(ctx context.Context, requestUUID uuid.UUID) (chan string, error) {
	ch, err := s.SSEBroker.GetOrCreateClient(requestUUID)
	return ch, err
}

func (s *TaskService) CloseSSE(ctx context.Context, requestUUID uuid.UUID) {
	s.SSEBroker.CloseClient(requestUUID)
}
