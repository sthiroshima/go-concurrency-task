package handlers

import (
	"encoding/json"
	"go-concurrency-task/internal/dto"
	"go-concurrency-task/internal/service"
	"net/http"

	"github.com/google/uuid"
)

type TaskHandler struct {
	TaskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{TaskService: taskService}
}

func (h *TaskHandler) Tasks(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// make checkers on valid request

	var res []*dto.TaskResponse

	res, err := h.TaskService.GetAllTasks()

	if err != nil {
		// TODO:
		// make headers code and response json
		return
	}

	_ = json.NewEncoder(w).Encode(&res)
}

func (h TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// make checkers on valid request
	var req *dto.TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return
	}

	var res *dto.TaskResponse
	res, err := h.TaskService.CreateTask(req)
	if err != nil {
		// TODO:
		// make headers code and response json
		return
	}

	_ = json.NewEncoder(w).Encode(&res)
}

func (h TaskHandler) TaskById(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// make checkers on valid request

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		// TODO:
		// make headers code and response json
		return
	}

	res, err := h.TaskService.TaskById(id)
	// TODO:
	// make checkers on valid request
	if err != nil {
		// TODO:
		// make headers code and response json
		return
	}

	_ = json.NewEncoder(w).Encode(&res)
}

func (h TaskHandler) DeleteTaskById(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// make checkers on valid request

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		// TODO:
		// make headers code and response json
		return
	}

	res, err := h.TaskService.DeleteTaskById(id)
	if err != nil {
		// TODO:
		// make headers code and response json
		return
	}

	_ = json.NewEncoder(w).Encode(&res)
}

func (h TaskHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// make checkers on valid request
	_, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		// TODO:
		// make headers code and response json
		return
	}

	return
}
