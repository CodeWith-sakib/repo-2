package events

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

// BusSubscriber receives events from subscribed topics.
type BusSubscriber struct {
	ID         string
	Topic      string
	Ch         chan *core.Event
	cancel     context.CancelFunc
	BufferSize int
}

// BusMetrics holds runtime statistics of the event bus.
type BusMetrics struct {
	Published  atomic.Int64
	Delivered  atomic.Int64
	Dropped    atomic.Int64
	DeadLetter atomic.Int64
}

// EventBus manages asynchronous fan-out message distribution across topics.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]*BusSubscriber // topic -> subscribers
	deadLetter  chan *core.Event
	metrics     BusMetrics
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	closed      bool
}

// NewEventBus creates a new running EventBus instance.
func NewEventBus(deadLetterCapacity int) *EventBus {
	if deadLetterCapacity <= 0 {
		deadLetterCapacity = 1000
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &EventBus{
		subscribers: make(map[string][]*BusSubscriber),
		deadLetter:  make(chan *core.Event, deadLetterCapacity),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Subscribe registers a subscriber to a topic with a specified buffer size.
func (b *EventBus) Subscribe(topic string, bufferSize int) (*BusSubscriber, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, fmt.Errorf("event bus is closed")
	}

	if bufferSize <= 0 {
		bufferSize = 64
	}

	_, subCancel := context.WithCancel(b.ctx)
	sub := &BusSubscriber{
		ID:         fmt.Sprintf("sub-%s-%d", topic, time.Now().UnixNano()),
		Topic:      topic,
		Ch:         make(chan *core.Event, bufferSize),
		cancel:     subCancel,
		BufferSize: bufferSize,
	}

	b.subscribers[topic] = append(b.subscribers[topic], sub)
	return sub, nil
}

// Unsubscribe removes a subscriber from the bus.
func (b *EventBus) Unsubscribe(sub *BusSubscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs := b.subscribers[sub.Topic]
	for i, s := range subs {
		if s.ID == sub.ID {
			b.subscribers[sub.Topic] = append(subs[:i], subs[i+1:]...)
			s.cancel()
			close(s.Ch)
			break
		}
	}
}

// Publish dispatches an event to all subscribers of the event topic.
func (b *EventBus) Publish(event *core.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return fmt.Errorf("event bus is closed")
	}

	b.metrics.Published.Add(1)
	topic := string(event.Type)
	subs := b.subscribers[topic]

	if len(subs) == 0 {
		// Route to dead letter queue
		select {
		case b.deadLetter <- event:
			b.metrics.DeadLetter.Add(1)
		default:
			b.metrics.Dropped.Add(1)
		}
		return nil
	}

	for _, sub := range subs {
		select {
		case sub.Ch <- event:
			b.metrics.Delivered.Add(1)
		default:
			// Buffer full, forward to dead letter queue
			b.metrics.Dropped.Add(1)
			select {
			case b.deadLetter <- event:
				b.metrics.DeadLetter.Add(1)
			default:
				// DLQ also full
			}
		}
	}

	return nil
}

// DeadLetterChannel returns the read-only channel for dead-lettered events.
func (b *EventBus) DeadLetterChannel() <-chan *core.Event {
	return b.deadLetter
}

// Metrics returns a snapshot of bus traffic metrics.
func (b *EventBus) Metrics() (published, delivered, dropped, deadLetter int64) {
	return b.metrics.Published.Load(),
		b.metrics.Delivered.Load(),
		b.metrics.Dropped.Load(),
		b.metrics.DeadLetter.Load()
}

// Close gracefully stops the event bus and cleans up subscribers.
func (b *EventBus) Close() {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	b.closed = true
	b.cancel()

	for _, subs := range b.subscribers {
		for _, s := range subs {
			s.cancel()
			close(s.Ch)
		}
	}
	b.subscribers = make(map[string][]*BusSubscriber)
	close(b.deadLetter)
	b.mu.Unlock()
}
