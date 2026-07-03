package main

import (
	"context"
	"go-concurrency-task/internal/repository"
	"go-concurrency-task/internal/service"
	"go-concurrency-task/internal/transport"
	"go-concurrency-task/internal/transport/handlers"
	"go-concurrency-task/internal/workers"
	"log"
	"net/http"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	sseBroker := workers.NewBroker()

	repo := repository.NewTaskStateRepository()
	queue := workers.NewTaskManager(ctx, repo, sseBroker)

	taskService := service.NewTaskService(repo, queue, sseBroker)
	handler := handlers.NewTaskHandler(taskService)

	server := http.Server{
		Addr:    ":8050",
		Handler: transport.NewRouter(handler),
	}
	s := transport.App{
		Server: &server,
	}

	go queue.Run(ctx)
	go sseBroker.Run(ctx)

	if err := s.Run(ctx); err != nil {
		log.Fatal(err)
	}

}
