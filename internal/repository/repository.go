package repository

import (
	"errors"
	"fmt"
	"go-concurrency-task/internal/domain"
	"sync"

	"github.com/google/uuid"
)

type TaskStateRepository struct {
	rwmu  sync.RWMutex
	tasks map[uuid.UUID]*domain.TaskState
}

func NewTaskStateRepository() *TaskStateRepository {
	return &TaskStateRepository{
		tasks: make(map[uuid.UUID]*domain.TaskState),
	}
}

func (s *TaskStateRepository) GetTasks() (map[uuid.UUID]domain.TaskState, error) {
	s.rwmu.RLock()
	defer s.rwmu.RUnlock()

	res := map[uuid.UUID]domain.TaskState{}
	for id, state := range s.tasks {
		res[id] = *state
	}

	return res, nil
}

func (s *TaskStateRepository) getById(ID uuid.UUID) (*domain.TaskState, error) {
	if ts, ok := s.tasks[ID]; ok {
		return ts, nil
	}

	return nil, errors.New("task not found")
}

func (s *TaskStateRepository) Add(taskState *domain.TaskState) error {
	s.rwmu.Lock()
	defer s.rwmu.Unlock()

	ID := taskState.Task.ID
	if _, ok := s.tasks[ID]; ok {
		return fmt.Errorf("task %s already exists", ID)
	}

	s.tasks[ID] = taskState

	return nil
}

func (s *TaskStateRepository) GetTaskState(ID uuid.UUID) (domain.TaskState, error) {
	s.rwmu.RLock()
	defer s.rwmu.RUnlock()

	taskState, err := s.getById(ID)
	if err != nil {
		return domain.TaskState{}, err
	}

	return *taskState, nil
}

func (s *TaskStateRepository) CancelTask(ID uuid.UUID) error {
	s.rwmu.Lock()
	defer s.rwmu.Unlock()

	taskState, err := s.getById(ID)
	if err != nil {
		return err
	}

	return taskState.MarkCancel()
}

func (s *TaskStateRepository) DoneTask(ID uuid.UUID) error {
	s.rwmu.Lock()
	defer s.rwmu.Unlock()

	taskState, err := s.getById(ID)
	if err != nil {
		return err
	}

	return taskState.MarkDone()
}

func (s *TaskStateRepository) ProcessingTask(ID uuid.UUID) error {
	s.rwmu.Lock()
	defer s.rwmu.Unlock()

	taskState, err := s.getById(ID)
	if err != nil {
		return err
	}

	return taskState.MarkProcessing()
}

func (s *TaskStateRepository) RetryProcessingTask(ID uuid.UUID) error {
	s.rwmu.Lock()
	defer s.rwmu.Unlock()

	taskState, err := s.getById(ID)
	if err != nil {
		return err
	}

	return taskState.MarkRetryProcessing()
}

func (s *TaskStateRepository) FailedTask(ID uuid.UUID) error {
	s.rwmu.Lock()
	defer s.rwmu.Unlock()

	taskState, err := s.getById(ID)
	if err != nil {
		return err
	}

	return taskState.MarkFailed()
}
