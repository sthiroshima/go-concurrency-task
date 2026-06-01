package transport

import (
	"context"
	"net/http"
)

type App struct {
	Server *http.Server
}

func (s *App) Run(ctx context.Context) error {
	// make run in goroutine
	if err := s.Server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
