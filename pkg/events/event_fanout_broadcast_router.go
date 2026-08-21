package events

import (
	"context"
	"sync"
)

// FanoutDestination represents a sink channel or subscriber queue for broadcasts.
type FanoutDestination struct {
	ID       string `json:"id"`
	QueueLen int    `json:"queue_len"`
}

// FanoutBroadcastRouter distributes incoming events concurrently across multiple subscriber streams.
type FanoutBroadcastRouter struct {
	mu           sync.RWMutex
	destinations map[string][]FanoutDestination // topic -> destinations
}

// NewFanoutBroadcastRouter creates a broadcast dispatcher.
func NewFanoutBroadcastRouter() *FanoutBroadcastRouter {
	return &FanoutBroadcastRouter{
		destinations: make(map[string][]FanoutDestination),
	}
}

// RegisterDestination attaches a listener to a topic fanout group.
func (r *FanoutBroadcastRouter) RegisterDestination(topic string, dest FanoutDestination) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.destinations[topic] = append(r.destinations[topic], dest)
}

// RouteFanout calculates destinations that will receive the broadcast.
func (r *FanoutBroadcastRouter) RouteFanout(ctx context.Context, topic string) []FanoutDestination {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dests := r.destinations[topic]
	out := make([]FanoutDestination, len(dests))
	copy(out, dests)
	return out
}

// DestinationCount returns number of subscribers bound to topic.
func (r *FanoutBroadcastRouter) DestinationCount(topic string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.destinations[topic])
}
