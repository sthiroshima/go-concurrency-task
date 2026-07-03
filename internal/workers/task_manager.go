package workers

import (
	"context"
	"fmt"
	"go-concurrency-task/internal/repository"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TaskManager struct {
	appCtx    context.Context
	repo      *repository.TaskStateRepository
	sseBroker *Broker
	queue     chan TaskCtx
	result    chan TaskResult
	attempts  map[uuid.UUID]int
	CancelCtx map[uuid.UUID]context.CancelFunc
	mu        sync.RWMutex
	wg        sync.WaitGroup
}

type TaskCtx struct {
	ID  uuid.UUID
	ctx context.Context
}

type TaskResult struct {
	ID     uuid.UUID
	Result bool
}

func NewTaskManager(appCtx context.Context, repo *repository.TaskStateRepository, sseBroker *Broker) *TaskManager {
	return &TaskManager{
		appCtx:    appCtx,
		repo:      repo,
		sseBroker: sseBroker,
		queue:     make(chan TaskCtx, 100),
		attempts:  make(map[uuid.UUID]int),
		result:    make(chan TaskResult, 100),
		CancelCtx: make(map[uuid.UUID]context.CancelFunc),
	}
}

func (m *TaskManager) Run(ctx context.Context) {
	for i := 0; i < 4; i++ {
		m.wg.Add(1)
		go m.worker(ctx)
	}

	m.wg.Add(1)
	go m.resultCollector(ctx)
}

func (m *TaskManager) worker(appCtx context.Context) {
	defer m.wg.Done()

	for {
		select {
		case task := <-m.queue:
			fmt.Printf("from queue %s\n", task.ID)
			m.executeTask(task.ctx, task.ID)
		case <-appCtx.Done():
			return
		}
	}
}

func (m *TaskManager) executeTask(ctx context.Context, ID uuid.UUID) {
	if err := m.repo.ProcessingTask(ID); err != nil {
		fmt.Println(err.Error())
		return
	}
	m.sseBroker.WriteMessage(fmt.Sprintf("%v is processing", ID))

	var jobResult bool
	for i := 0; i < 20; i++ {
		select {
		case <-ctx.Done():
			goto write
		case <-time.After(time.Second * time.Duration(rand.Intn(2))):
			if rand.Intn(20) == 2 {
				jobResult = true
				goto write
			}
		}
	}

write:
	m.result <- TaskResult{ID: ID, Result: jobResult}
	m.ClearCancelCtx(ID)
}

func (m *TaskManager) resultCollector(ctx context.Context) {
	defer m.wg.Done()

	for {
		select {
		case res := <-m.result:
			if m.retryTask(res.ID) {
				continue
			}

			m.saveResult(ctx, res)
		case <-ctx.Done():
			return
		}
	}
}

func (m *TaskManager) saveResult(ctx context.Context, res TaskResult) {
	switch res.Result {
	case true:
		if err := m.repo.DoneTask(res.ID); err != nil {
			fmt.Println(err.Error())
			return
		}
		m.sseBroker.WriteMessage(fmt.Sprintf("%v is done", res.ID))
	case false:
		if err := m.repo.FailedTask(res.ID); err != nil {
			fmt.Println(err.Error())
			return
		}
		m.sseBroker.WriteMessage(fmt.Sprintf("%v is failed", res.ID))
	}
}

func (m *TaskManager) retryTask(ID uuid.UUID) bool {
	if m.addAttempt(ID) {
		if err := m.repo.RetryProcessingTask(ID); err != nil {
			panic(err)
			return false
		}
		m.sseBroker.WriteMessage(fmt.Sprintf("%v is retry", ID))

		m.AdToQueue(ID)
		return true
	}

	return false
}

func (m *TaskManager) addAttempt(ID uuid.UUID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.attempts[ID] >= 2 {
		delete(m.attempts, ID)
		return false

	}

	m.attempts[ID]++
	return true
}

func (m *TaskManager) AdToQueue(ID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithCancel(m.appCtx)
	task := TaskCtx{
		ID:  ID,
		ctx: ctx,
	}
	m.CancelCtx[ID] = cancel
	m.queue <- task
}

func (m *TaskManager) ClearCancelCtx(id uuid.UUID) {
	m.mu.Lock()
	delete(m.CancelCtx, id)
	m.mu.Unlock()
}

func (m *TaskManager) CancelTask(ID uuid.UUID) error {
	m.mu.RLock()
	cancelFunc, ok := m.CancelCtx[ID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("task id %v not found", ID)
	}

	cancelFunc()
	m.ClearCancelCtx(ID)
	return nil
}
