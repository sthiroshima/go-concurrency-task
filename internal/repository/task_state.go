package repository

import "go-concurrency-task/internal/domain"

type TaskStateRepository struct {
	tasks map[string]*domain.TaskState
}

// TODO:
// read and write with sync.RWMutex
