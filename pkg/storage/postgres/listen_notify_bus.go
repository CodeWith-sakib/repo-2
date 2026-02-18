package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// NotificationEvent captures a PostgreSQL LISTEN/NOTIFY asynchronous message.
type NotificationEvent struct {
	Channel   string          `json:"channel"`
	Payload   json.RawMessage `json:"payload"`
	ReceivedAt time.Time      `json:"received_at"`
}

// NotificationHandler is invoked upon receiving a live database notification.
type NotificationHandler func(event NotificationEvent)

// ListenNotifyBus manages simulated or driver-backed asynchronous pub/sub channels.
type ListenNotifyBus struct {
	mu       sync.RWMutex
	handlers map[string][]NotificationHandler
	db       *sql.DB
}

// NewListenNotifyBus creates a PostgreSQL listen/notify bus.
func NewListenNotifyBus(db *sql.DB) *ListenNotifyBus {
	return &ListenNotifyBus{
		handlers: make(map[string][]NotificationHandler),
		db:       db,
	}
}

// Subscribe registers a listener callback for a named notification channel.
func (b *ListenNotifyBus) Subscribe(channel string, handler NotificationHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[channel] = append(b.handlers[channel], handler)
}

// Notify publishes a notification payload to channel via PG pg_notify or in-memory dispatch.
func (b *ListenNotifyBus) Notify(ctx context.Context, channel string, payload interface{}) error {
	if channel == "" {
		return errors.New("channel cannot be empty")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed marshaling notification payload: %w", err)
	}

	if b.db != nil {
		query := "SELECT pg_notify($1, $2)"
		if _, err := b.db.ExecContext(ctx, query, channel, string(data)); err != nil {
			return fmt.Errorf("failed issuing pg_notify: %w", err)
		}
	}

	// Dispatch to registered local in-process listeners
	b.mu.RLock()
	handlers := append([]NotificationHandler(nil), b.handlers[channel]...)
	b.mu.RUnlock()

	evt := NotificationEvent{
		Channel:    channel,
		Payload:    data,
		ReceivedAt: time.Now().UTC(),
	}

	for _, h := range handlers {
		h(evt)
	}

	return nil
}
