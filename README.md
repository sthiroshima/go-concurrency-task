# go-concurrency-task

A lightweight backend service written in Go to deeply explore concurrency concepts. The project implements an HTTP API that accepts tasks, queues them, and processes them asynchronously using a worker pool built with goroutines and channels.

## Features

- **HTTP API** for creating, listing, fetching, and canceling tasks
- **Worker pool** with 4 concurrent goroutines processing tasks from a buffered channel
- **Task lifecycle** with explicit status transitions (`queued` → `processing` → `done` / `failed`)
- **Automatic retries** — failed tasks are retried up to 2 times before being marked as failed
- **Graceful cancellation** — queued tasks can be canceled via `DELETE`; in-flight tasks are stopped via `context.CancelFunc`
- **SSE (Server-Sent Events)** — real-time task status updates streamed to connected clients
- **In-memory storage** with `sync.RWMutex` for thread-safe access

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌──────────────────┐
│   HTTP API  │────▶│   Service   │────▶│   Repository     │
│  (transport)│     │             │     │  (in-memory)     │
└─────────────┘     └──────┬──────┘     └──────────────────┘
                           │
                    ┌──────▼──────┐     ┌──────────────────┐
                    │ TaskManager │────▶│   SSE Broker     │
                    │ (4 workers) │     │  (pub/sub)       │
                    └─────────────┘     └──────────────────┘
```

### Concurrency primitives used

| Component | Mechanism |
|-----------|-----------|
| `TaskManager` | Buffered channels (`queue`, `result`), `sync.WaitGroup`, `sync.RWMutex`, `context.Context` |
| `TaskStateRepository` | `sync.RWMutex` for safe concurrent reads/writes |
| `Broker` (SSE) | Fan-out broadcast channel, per-client event channels, `sync.RWMutex` |
| Shutdown | `signal.NotifyContext` for `SIGINT` / `SIGTERM` |

## Project structure

```
.
├── cmd/
│   ├── main.go              # Application entry point
│   └── test/main.go         # Concurrency experiments (scoreboard pattern)
├── internal/
│   ├── domain/              # Task and TaskState entities, status transitions
│   ├── dto/                 # Request/response DTOs
│   ├── repository/          # In-memory task storage
│   ├── service/             # Business logic
│   ├── transport/           # HTTP server, router, handlers
│   └── workers/             # TaskManager (worker pool) and SSE Broker
├── docker/
│   └── Dockerfile
└── docker-compose.yml
```

## Task statuses

| Status | Description |
|--------|-------------|
| `queued` | Task created and waiting in the queue |
| `processing` | A worker is executing the task |
| `retryProcessing` | Task failed and is being retried |
| `done` | Task completed successfully |
| `failed` | Task failed after all retry attempts |
| `canceled` | Task was canceled while still in `queued` state |

## API

Base URL: `http://localhost:8050`

### `POST /tasks` — Create a task

**Request:**
```json
{
  "type": "email",
  "payload": "send welcome email"
}
```

**Response:**
```json
{
  "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "status": "queued"
}
```

### `GET /tasks` — List all tasks

**Response:**
```json
[
  {
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "status": "done"
  }
]
```

### `GET /tasks/{id}` — Get task by ID

**Response:**
```json
{
  "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "status": "processing"
}
```

### `DELETE /tasks/{id}` — Cancel a task

Cancels a task that is still in `queued` status, or stops an in-flight task via context cancellation.

**Response:**
```json
{
  "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "status": "canceled"
}
```

### `GET /events` — SSE stream

Opens a Server-Sent Events connection to receive real-time task updates:

```
data: 3fa85f64-5717-4562-b3fc-2c963f66afa6 is processing

data: 3fa85f64-5717-4562-b3fc-2c963f66afa6 is done
```

### `GET /metrics` — Metrics (WIP)

Endpoint is registered but not yet implemented.

## Getting started

### Prerequisites

- Go 1.26+
- Docker & Docker Compose (optional)

### Run locally

```bash
go run ./cmd
```

The server starts on port `8050`.

### Run with Docker

```bash
docker compose up --build
```

The service is available at `http://localhost:8050`.

### Example usage

```bash
# Create a task
curl -X POST http://localhost:8050/tasks \
  -H "Content-Type: application/json" \
  -d '{"type": "email", "payload": "hello"}'

# List all tasks
curl http://localhost:8050/tasks

# Get task by ID
curl http://localhost:8050/tasks/<task-id>

# Cancel a task
curl -X DELETE http://localhost:8050/tasks/<task-id>

# Subscribe to SSE events
curl -N http://localhost:8050/events
```

## How task processing works

1. A new task is saved to the repository with status `queued` and pushed to the worker queue.
2. One of 4 workers picks up the task, sets status to `processing`, and simulates work (random delay, ~5% success chance per tick).
3. On success → status `done`; on failure → up to 2 retries with status `retryProcessing`, then `failed`.
4. Each status change is broadcast to all SSE clients via the Broker.

## Dependencies

- [github.com/google/uuid](https://github.com/google/uuid) — UUID generation
