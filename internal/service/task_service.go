package service

import (
	"go-concurrency-task/internal/domain"
	"go-concurrency-task/internal/repository"
)

type TaskService struct {
	taskRepo *repository.TaskStateRepository
	queue    chan *domain.Task
}
