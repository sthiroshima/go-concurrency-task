package transport

import (
	"go-concurrency-task/internal/transport/handlers"
	"net/http"
)

func NewRouter(taskHandler *handlers.TaskHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", taskHandler.CreateTask)
	mux.HandleFunc("GET /tasks/{id}", taskHandler.GetTask)
	return mux
}
