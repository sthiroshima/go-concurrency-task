package dto

import "github.com/google/uuid"

type TaskRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type TaskResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}
