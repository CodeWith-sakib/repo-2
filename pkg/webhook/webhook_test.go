package webhook

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestWebhookDeliveryWithHMAC(t *testing.T) {
	secret := "test-secret-key-1234"
	var receivedSig string
	var receivedEvent string

	mockClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			receivedSig = req.Header.Get("X-Kestrel-Signature")
			receivedEvent = req.Header.Get("X-Kestrel-Event")
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"status":"received"}`))),
				Header:     make(http.Header),
			}, nil
		}),
	}

	dispatcher := NewDispatcher(mockClient)
	event := &core.Event{
		ID:        core.NewID("evt"),
		Type:      core.EventRunCompleted,
		TenantID:  "tenant-webhook",
		RunID:     core.NewID("run"),
		Timestamp: time.Now().UTC(),
	}

	target := Target{
		ID:        "target-1",
		URL:       "https://webhook.internal/events",
		SecretKey: secret,
	}

	if err := dispatcher.Dispatch(context.Background(), target, event); err != nil {
		t.Fatalf("webhook dispatch failed: %v", err)
	}

	if receivedEvent != string(core.EventRunCompleted) {
		t.Errorf("expected event %s, got %s", core.EventRunCompleted, receivedEvent)
	}
	if receivedSig == "" {
		t.Error("expected HMAC signature in request, got empty")
	}

	// Verify dead letter queue on 500 error
	errorClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewReader([]byte(`server error`))),
				Header:     make(http.Header),
			}, nil
		}),
	}
	errDispatcher := NewDispatcher(errorClient)
	_ = errDispatcher.Dispatch(context.Background(), target, event)

	dls := errDispatcher.DeadLetters()
	if len(dls) != 1 {
		t.Fatalf("expected 1 dead letter after failure, got %d", len(dls))
	}
	if dls[0].TargetID != "target-1" {
		t.Errorf("dead letter target mismatch: %s", dls[0].TargetID)
	}
}
