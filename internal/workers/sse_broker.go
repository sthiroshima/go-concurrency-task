package workers

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type Client struct {
	events chan string
	closed bool
}

type Broker struct {
	broadcast chan string
	closed    bool
	mu        sync.RWMutex
	wg        sync.WaitGroup
	clients   map[uuid.UUID]*Client
}

func NewBroker() *Broker {
	return &Broker{
		broadcast: make(chan string, 100),
		closed:    false,
		clients:   make(map[uuid.UUID]*Client),
	}
}

func NewClient() *Client {
	c := Client{
		events: make(chan string, 100),
		closed: false,
	}
	return &c
}

func (b *Broker) Run(ctx context.Context) {
	b.wg.Add(1)

	go b.worker(ctx)
}

func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.closed {
		close(b.broadcast)
	}

	b.closed = true

	for UUID, client := range b.clients {
		if !client.closed {
			b.clients[UUID].closed = true
			close(client.events)
		}

		delete(b.clients, UUID)
	}
}

func (b *Broker) worker(appCtx context.Context) {
	defer b.wg.Done()

	for {
		select {
		case message, ok := <-b.broadcast:
			if !ok {
				return
			}

			b.mu.RLock()
			for _, client := range b.clients {
				if client.closed {
					continue
				}

				select {
				case client.events <- message:
				default:
				}
			}
			b.mu.RUnlock()
		case <-appCtx.Done():
			b.Close()
			b.wg.Wait()

			return
		}
	}
}

func (b *Broker) WriteMessage(ctx context.Context, message string) {
	if b.closed {
		return
	}

	select {
	case <-ctx.Done():
	case b.broadcast <- message:
	default:
	}
}

func (b *Broker) GetOrCreateClient(requestUUID uuid.UUID) chan string {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	c, ok := b.clients[requestUUID]
	if !ok {
		c = NewClient()
		b.clients[requestUUID] = c
	}

	return c.events
}

func (b *Broker) CloseClient(requestUUID uuid.UUID) {
	b.mu.Lock()
	defer b.mu.Unlock()

	c, ok := b.clients[requestUUID]
	if ok {
		if !c.closed {
			c.closed = true
			close(c.events)
		}

		delete(b.clients, requestUUID)
	}
}
