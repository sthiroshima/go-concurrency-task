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
	dispatcher chan string
	rwmu       sync.RWMutex
	wg         sync.WaitGroup
	clients    map[uuid.UUID]*Client
}

func NewBroker() *Broker {
	return &Broker{
		dispatcher: make(chan string, 100),
		clients:    make(map[uuid.UUID]*Client),
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

func (b *Broker) worker(appCtx context.Context) {
	defer b.wg.Done()

	for {
		select {
		case message := <-b.dispatcher:
			b.rwmu.RLock()
			for _, client := range b.clients {
				select {
				case client.events <- message:
				default:
				}
			}
			b.rwmu.RUnlock()
		case <-appCtx.Done():
			return
		}
	}
}

func (b *Broker) WriteMessage(message string) {
	select {
	case b.dispatcher <- message:
	default:
	}
}

func (b *Broker) GetOrCreateClient(requestUUID uuid.UUID) (chan string, error) {
	c, ok := b.clients[requestUUID]
	if !ok {
		c = NewClient()

		b.rwmu.Lock()
		b.clients[requestUUID] = c
		b.rwmu.Unlock()
	}

	return c.events, nil
}

func (b *Broker) CloseClient(requestUUID uuid.UUID) {
	b.rwmu.Lock()
	defer b.rwmu.Unlock()

	c, ok := b.clients[requestUUID]
	if ok {
		close(c.events)
	}

	delete(b.clients, requestUUID)
}
