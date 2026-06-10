package transport

import (
	"go-concurrency-task/internal/transport/handlers"
	"net/http"
)

func NewRouter(taskHandler *handlers.TaskHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks", taskHandler.Tasks)
	mux.HandleFunc("POST /tasks", taskHandler.CreateTask)
	mux.HandleFunc("GET /tasks/{id}", taskHandler.TaskById)
	mux.HandleFunc("DELETE /tasks/{id}", taskHandler.DeleteTaskById)
	mux.HandleFunc("GET /metrics", taskHandler.GetMetrics)
	return mux
}
