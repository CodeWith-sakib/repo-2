package webhook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestWebhookDeliveryWithHMAC(t *testing.T) {
	secret := "test-secret-key-1234"
	var receivedSig string
	var receivedEvent core.EventType

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSig = r.Header.Get("X-Kestrel-Signature")
		receivedEvent = core.EventType(r.Header.Get("X-Kestrel-Event"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := NewDispatcher(server.Client())
	event := &core.Event{
		ID:        core.NewID("evt"),
		Type:      core.EventRunCompleted,
		TenantID:  "tenant-webhook",
		RunID:     core.NewID("run"),
		Timestamp: time.Now().UTC(),
	}

	target := Target{
		ID:        "target-1",
		URL:       server.URL,
		SecretKey: secret,
	}

	if err := dispatcher.Dispatch(context.Background(), target, event); err != nil {
		t.Fatalf("webhook dispatch failed: %v", err)
	}

	if receivedEvent != core.EventRunCompleted {
		t.Errorf("expected event %s, got %s", core.EventRunCompleted, receivedEvent)
	}
	if receivedSig == "" {
		t.Error("expected HMAC signature in request, got empty")
	}
}
