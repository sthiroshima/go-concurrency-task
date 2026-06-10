package domain

import "fmt"

type Status string

const (
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusDone       Status = "done"
	StatusFailed     Status = "failed"
	StatusCanceled   Status = "canceled"
)

type TaskState struct {
	Task   Task
	Status Status
}

func NewTaskState(task *Task) *TaskState {
	return &TaskState{
		Task:   *task,
		Status: StatusQueued,
	}
}

func (s *TaskState) MarkDone() error {
	if s.Status == StatusProcessing {
		s.Status = StatusDone
		return nil
	}

	return fmt.Errorf("Done only processing status, current status is %s", s.Status)
}

func (s *TaskState) MarkCancel() error {
	if s.Status == StatusQueued {
		s.Status = StatusCanceled
		return nil
	}

	return fmt.Errorf("Cancel only Queued status, current status is %s", s.Status)
}

func (s *TaskState) MarkFailed() error {
	if s.Status == StatusProcessing {
		s.Status = StatusFailed
		return nil
	}

	return fmt.Errorf("Failed only Processing status, current status is %s", s.Status)
}
func (s *TaskState) MarkProcessing() error {
	if s.Status == StatusQueued {
		s.Status = StatusProcessing
		return nil
	}

	return fmt.Errorf("Processing only queued status, current status is %s", s.Status)
}
