package events

import (
	"context"
	"sync"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type Handler func(ctx context.Context, event *core.Event) error

type Bus struct {
	mu          sync.RWMutex
	subscribers map[core.EventType][]Handler
	globalSubs  []Handler
}

func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[core.EventType][]Handler),
		globalSubs:  make([]Handler, 0),
	}
}

func (b *Bus) Subscribe(eventType core.EventType, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], h)
}

func (b *Bus) SubscribeAll(h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.globalSubs = append(b.globalSubs, h)
}

func (b *Bus) Publish(ctx context.Context, event *core.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlers := append([]Handler{}, b.globalSubs...)
	if specific, ok := b.subscribers[event.Type]; ok {
		handlers = append(handlers, specific...)
	}

	for _, h := range handlers {
		_ = h(ctx, event)
	}
}
