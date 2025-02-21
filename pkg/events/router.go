package events

import (
	"context"
	"strings"
	"sync"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type EventHandler func(ctx context.Context, event *core.Event) error

type EventRouter struct {
	mu       sync.RWMutex
	routes   map[string][]EventHandler
	wildcard []EventHandler
}

func NewEventRouter() *EventRouter {
	return &EventRouter{
		routes: make(map[string][]EventHandler),
	}
}

func (r *EventRouter) Subscribe(pattern string, handler EventHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if pattern == "*" || pattern == "#" {
		r.wildcard = append(r.wildcard, handler)
		return
	}

	r.routes[pattern] = append(r.routes[pattern], handler)
}

func (r *EventRouter) Dispatch(ctx context.Context, event *core.Event) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var handlers []EventHandler
	handlers = append(handlers, r.wildcard...)

	eventTypeStr := string(event.Type)
	if direct, exists := r.routes[eventTypeStr]; exists {
		handlers = append(handlers, direct...)
	}

	// Prefix matching like "step.*"
	parts := strings.Split(eventTypeStr, ".")
	if len(parts) == 2 {
		prefixPattern := parts[0] + ".*"
		if prefixHandlers, exists := r.routes[prefixPattern]; exists {
			handlers = append(handlers, prefixHandlers...)
		}
	}

	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
