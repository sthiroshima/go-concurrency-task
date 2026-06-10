package domain

import (
	"github.com/google/uuid"
)

type Task struct {
	ID      uuid.UUID `json:"id"`
	Type    string    `json:"type"`
	Payload string    `json:"payload"`
}

func NewTask(ID uuid.UUID, taskType string, payload string) *Task {
	return &Task{
		ID:      ID,
		Type:    taskType,
		Payload: payload,
	}
}
