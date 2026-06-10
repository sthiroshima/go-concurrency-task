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
	queue     chan TaskCtx
	result    chan TaskResult
	CancelCtx map[uuid.UUID]context.CancelFunc
	rwmu      sync.RWMutex
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

func NewTaskManager(appCtx context.Context, repo *repository.TaskStateRepository) *TaskManager {
	return &TaskManager{
		appCtx:    appCtx,
		repo:      repo,
		queue:     make(chan TaskCtx, 100),
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
			m.executeTask(task.ctx, task.ID)
		case <-appCtx.Done():
			return
		}
	}
}

func (m *TaskManager) executeTask(ctx context.Context, ID uuid.UUID) {
	if err := m.repo.ProcessingTask(ID); err != nil {
		fmt.Printf("mark is processing failed, %v\n", ID)
		return
	}

	var jobResult bool
	for i := 0; i < 20; i++ {
		select {
		case <-ctx.Done():
			goto write
		case <-time.After(time.Second * time.Duration(rand.Intn(15))):
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
	case false:
		if err := m.repo.FailedTask(res.ID); err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func (m *TaskManager) AdToQueue(ID uuid.UUID) {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	ctx, cancel := context.WithCancel(m.appCtx)
	task := TaskCtx{
		ID:  ID,
		ctx: ctx,
	}
	m.CancelCtx[ID] = cancel
	m.queue <- task
}

func (m *TaskManager) ClearCancelCtx(id uuid.UUID) {
	m.rwmu.Lock()
	delete(m.CancelCtx, id)
	m.rwmu.Unlock()
}

func (m *TaskManager) CancelTask(ID uuid.UUID) error {
	m.rwmu.RLock()
	cancelFunc, ok := m.CancelCtx[ID]
	m.rwmu.RUnlock()
	if !ok {
		return fmt.Errorf("task id %v not found", ID)
	}

	cancelFunc()
	m.ClearCancelCtx(ID)
	return nil
}
