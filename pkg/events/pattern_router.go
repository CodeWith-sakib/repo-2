package events

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

// PatternRouteAction is the handler invoked when a pattern route matches.
type PatternRouteAction func(event *core.Event) error

// PatternRoute is a compiled routing rule matching events by type regexp and optional tenant.
type PatternRoute struct {
	ID       string
	Pattern  *regexp.Regexp
	TenantID string
	Action   PatternRouteAction
	Priority int
}

// PatternRouter dispatches incoming events to registered regex-priority routes.
type PatternRouter struct {
	mu     sync.RWMutex
	routes []*PatternRoute
	nextID int
}

// NewPatternRouter initializes a pattern-based event router.
func NewPatternRouter() *PatternRouter {
	return &PatternRouter{}
}

// RegisterRoute adds a routing rule and returns the assigned route ID.
func (r *PatternRouter) RegisterRoute(pattern string, tenantID string, priority int, action PatternRouteAction) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid route pattern %q: %w", pattern, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	id := fmt.Sprintf("route-%d", r.nextID)

	route := &PatternRoute{
		ID:       id,
		Pattern:  re,
		TenantID: tenantID,
		Action:   action,
		Priority: priority,
	}

	inserted := false
	for i, existing := range r.routes {
		if priority > existing.Priority {
			tail := make([]*PatternRoute, len(r.routes[i:]))
			copy(tail, r.routes[i:])
			r.routes = append(r.routes[:i], route)
			r.routes = append(r.routes, tail...)
			inserted = true
			break
		}
	}
	if !inserted {
		r.routes = append(r.routes, route)
	}

	return id, nil
}

// RemoveRoute deregisters a route by ID.
func (r *PatternRouter) RemoveRoute(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, route := range r.routes {
		if route.ID == id {
			r.routes = append(r.routes[:i], r.routes[i+1:]...)
			return true
		}
	}
	return false
}

// Route dispatches an event to the highest-priority matching route.
func (r *PatternRouter) Route(event *core.Event) error {
	r.mu.RLock()
	routes := make([]*PatternRoute, len(r.routes))
	copy(routes, r.routes)
	r.mu.RUnlock()

	for _, route := range routes {
		if !route.Pattern.MatchString(string(event.Type)) {
			continue
		}
		if route.TenantID != "" && route.TenantID != event.TenantID {
			continue
		}
		return route.Action(event)
	}

	return fmt.Errorf("no pattern route matched event type %q for tenant %q", event.Type, event.TenantID)
}

// RouteCount returns the number of registered routes.
func (r *PatternRouter) RouteCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.routes)
}
