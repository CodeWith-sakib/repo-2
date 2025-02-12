package events

import (
	"strings"
	"sync"
)

type EventHandler func(event *EventEnvelope) error

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

func (r *EventRouter) Dispatch(event *EventEnvelope) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var handlers []EventHandler
	handlers = append(handlers, r.wildcard...)

	if direct, exists := r.routes[event.Type]; exists {
		handlers = append(handlers, direct...)
	}

	// Prefix matching like "step.*"
	parts := strings.Split(event.Type, ".")
	if len(parts) == 2 {
		prefixPattern := parts[0] + ".*"
		if prefixHandlers, exists := r.routes[prefixPattern]; exists {
			handlers = append(handlers, prefixHandlers...)
		}
	}

	for _, h := range handlers {
		if err := h(event); err != nil {
			return err
		}
	}
	return nil
}
