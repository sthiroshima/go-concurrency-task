package handlers

import (
	"encoding/json"
	"fmt"
	"go-concurrency-task/internal/domain"
	"net/http"
)

type TaskHandler struct {
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{}
}

func (h TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	task := domain.Task{
		ID:      "544",
		Type:    "test",
		Payload: "test - test",
	}

	_ = json.NewEncoder(w).Encode(task)
}

func (h TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	fmt.Println(id)
}
