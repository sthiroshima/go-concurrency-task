package main

import (
	"context"
	"go-concurrency-task/internal/transport"
	"go-concurrency-task/internal/transport/handlers"
	"log"
	"net/http"
	"os/signal"
	"syscall"
)

func main() {

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer stop()

	taskHandler := handlers.NewTaskHandler()

	server := http.Server{
		Addr:    ":8080",
		Handler: transport.NewRouter(taskHandler),
	}

	s := transport.App{
		Server: &server,
	}

	if err := s.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
