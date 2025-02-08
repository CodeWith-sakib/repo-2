package webhook

import (
	"fmt"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type DeliveryStatus string

const (
	StatusPending   DeliveryStatus = "PENDING"
	StatusInFlight  DeliveryStatus = "IN_FLIGHT"
	StatusDelivered DeliveryStatus = "DELIVERED"
	StatusFailed    DeliveryStatus = "FAILED"
)

type QueuedDelivery struct {
	ID          string         `json:"id"`
	URL         string         `json:"url"`
	Secret      string         `json:"secret"`
	Payload     []byte         `json:"payload"`
	Attempts    int            `json:"attempts"`
	MaxAttempts int            `json:"max_attempts"`
	NextTry     time.Time      `json:"next_try"`
	Status      DeliveryStatus `json:"status"`
	LastError   string         `json:"last_error,omitempty"`
}

type DeliveryQueue struct {
	mu    sync.Mutex
	items map[string]*QueuedDelivery
}

func NewDeliveryQueue() *DeliveryQueue {
	return &DeliveryQueue{
		items: make(map[string]*QueuedDelivery),
	}
}

func (q *DeliveryQueue) Enqueue(id string, url, secret string, payload []byte, maxAttempts int) *QueuedDelivery {
	q.mu.Lock()
	defer q.mu.Unlock()

	item := &QueuedDelivery{
		ID:          id,
		URL:         url,
		Secret:      secret,
		Payload:     payload,
		MaxAttempts: maxAttempts,
		NextTry:     time.Now().UTC(),
		Status:      StatusPending,
	}
	q.items[id] = item
	return item
}

func (q *DeliveryQueue) GetDueDeliveries(now time.Time, limit int) []*QueuedDelivery {
	q.mu.Lock()
	defer q.mu.Unlock()

	var due []*QueuedDelivery
	for _, it := range q.items {
		if it.Status == StatusPending && !it.NextTry.After(now) {
			it.Status = StatusInFlight
			due = append(due, it)
			if limit > 0 && len(due) >= limit {
				break
			}
		}
	}
	return due
}

func (q *DeliveryQueue) MarkSuccess(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	it, ok := q.items[id]
	if !ok {
		return fmt.Errorf("%w: queued delivery not found", core.ErrNotFound)
	}
	it.Status = StatusDelivered
	return nil
}

func (q *DeliveryQueue) MarkFailed(id string, errReason string, backoff time.Duration) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	it, ok := q.items[id]
	if !ok {
		return fmt.Errorf("%w: queued delivery not found", core.ErrNotFound)
	}

	it.Attempts++
	it.LastError = errReason
	if it.Attempts >= it.MaxAttempts {
		it.Status = StatusFailed
	} else {
		it.Status = StatusPending
		it.NextTry = time.Now().Add(backoff)
	}
	return nil
}
