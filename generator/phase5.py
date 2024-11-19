import os
from generator.git_utils import commit
from generator.loc import get_production_loc

def write_file(path, content):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w', encoding='utf-8') as f:
        f.write(content.strip() + '\n')

def run_phase_5(dates_iter):
    print("=== Executing Phase 5: Plugin Architecture, Webhook & Event Serialization ===")

    # Commit 5.1: Plugin abstraction interface and handler registry
    write_file("pkg/plugin/plugin.go", """package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

type TaskHandler interface {
	Type() string
	Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error)
	ValidateConfig(config json.RawMessage) error
}

type Registry struct {
	mu       sync.RWMutex
	handlers map[string]TaskHandler
}

func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]TaskHandler),
	}
}

func (r *Registry) Register(h TaskHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	taskType := h.Type()
	if taskType == "" {
		return fmt.Errorf("plugin task type cannot be empty")
	}
	if _, exists := r.handlers[taskType]; exists {
		return fmt.Errorf("plugin already registered for type: %s", taskType)
	}

	r.handlers[taskType] = h
	return nil
}

func (r *Registry) Get(taskType string) (TaskHandler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	h, exists := r.handlers[taskType]
	if !exists {
		return nil, fmt.Errorf("no plugin registered for task type: %s", taskType)
	}
	return h, nil
}

func (r *Registry) PopulateWorkerRegistry(wreg *worker.ExecutorRegistry) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for k, h := range r.handlers {
		wreg.Register(k, h)
	}
}
""")
    commit("plugin: define TaskHandler interface and thread-safe plugin registry", next(dates_iter), [
        "pkg/plugin/plugin.go"
    ])

    # Commit 5.2: First-party HTTP Task Plugin
    write_file("pkg/plugin/builtin/http/http_plugin.go", """package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

type HTTPTaskConfig struct {
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers,omitempty"`
	Body           string            `json:"body,omitempty"`
	Timeout        string            `json:"timeout,omitempty"`
	ExpectedStatus int               `json:"expected_status,omitempty"`
}

type Plugin struct {
	client *http.Client
}

func NewPlugin(client *http.Client) *Plugin {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	return &Plugin{client: client}
}

func (p *Plugin) Type() string {
	return "http"
}

func (p *Plugin) ValidateConfig(cfgJSON json.RawMessage) error {
	if len(cfgJSON) == 0 {
		return fmt.Errorf("%w: http task config is required", core.ErrValidationFailed)
	}
	var cfg HTTPTaskConfig
	if err := json.Unmarshal(cfgJSON, &cfg); err != nil {
		return fmt.Errorf("%w: invalid http task config: %v", core.ErrValidationFailed, err)
	}
	if cfg.URL == "" {
		return fmt.Errorf("%w: http url is required", core.ErrValidationFailed)
	}
	if cfg.Method == "" {
		return fmt.Errorf("%w: http method is required", core.ErrValidationFailed)
	}
	return nil
}

func (p *Plugin) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	if err := p.ValidateConfig(sctx.Input); err != nil {
		return nil, err
	}

	var cfg HTTPTaskConfig
	_ = json.Unmarshal(sctx.Input, &cfg)

	timeout := 30 * time.Second
	if cfg.Timeout != "" {
		if d, err := time.ParseDuration(cfg.Timeout); err == nil && d > 0 {
			timeout = d
		}
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var bodyReader io.Reader
	if cfg.Body != "" {
		bodyReader = bytes.NewBufferString(cfg.Body)
	}

	req, err := http.NewRequestWithContext(reqCtx, cfg.Method, cfg.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading http response body: %w", err)
	}

	expectedStatus := cfg.ExpectedStatus
	if expectedStatus == 0 {
		expectedStatus = http.StatusOK
	}

	if resp.StatusCode != expectedStatus {
		return &worker.StepResult{
			ErrorMessage: fmt.Sprintf("unexpected status %d, expected %d: %s", resp.StatusCode, expectedStatus, string(respBody)),
			Retryable:    resp.StatusCode >= 500,
		}, nil
	}

	outputMap := map[string]interface{}{
		"status": resp.StatusCode,
		"body":   string(respBody),
	}
	outputJSON, _ := json.Marshal(outputMap)

	return &worker.StepResult{
		Output: outputJSON,
	}, nil
}
""")
    commit("plugin/http: implement first-party HTTP task execution plugin", next(dates_iter), [
        "pkg/plugin/builtin/http/http_plugin.go"
    ])

    # Commit 5.3: First-party Shell Task Plugin
    write_file("pkg/plugin/builtin/shell/shell_plugin.go", """package shell

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

type ShellTaskConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Timeout string            `json:"timeout,omitempty"`
}

type Plugin struct{}

func NewPlugin() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Type() string {
	return "shell"
}

func (p *Plugin) ValidateConfig(cfgJSON json.RawMessage) error {
	if len(cfgJSON) == 0 {
		return fmt.Errorf("%w: shell task config is required", core.ErrValidationFailed)
	}
	var cfg ShellTaskConfig
	if err := json.Unmarshal(cfgJSON, &cfg); err != nil {
		return fmt.Errorf("%w: invalid shell task config: %v", core.ErrValidationFailed, err)
	}
	if cfg.Command == "" {
		return fmt.Errorf("%w: shell command is required", core.ErrValidationFailed)
	}
	return nil
}

func (p *Plugin) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	if err := p.ValidateConfig(sctx.Input); err != nil {
		return nil, err
	}

	var cfg ShellTaskConfig
	_ = json.Unmarshal(sctx.Input, &cfg)

	timeout := 60 * time.Second
	if cfg.Timeout != "" {
		if d, err := time.ParseDuration(cfg.Timeout); err == nil && d > 0 {
			timeout = d
		}
	}

	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, cfg.Command, cfg.Args...)

	// Inject safe environment variables
	for k, v := range cfg.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("KESTREL_RUN_ID=%s", sctx.RunID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("KESTREL_STEP_ID=%s", sctx.StepID))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("shell execution failure: %w", err)
		}
	}

	resultMap := map[string]interface{}{
		"exit_code": exitCode,
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
	}
	outputJSON, _ := json.Marshal(resultMap)

	if exitCode != 0 {
		return &worker.StepResult{
			Output:       outputJSON,
			ErrorMessage: fmt.Sprintf("command exited with code %d: %s", exitCode, stderr.String()),
			Retryable:    false,
		}, nil
	}

	return &worker.StepResult{
		Output: outputJSON,
	}, nil
}
""")
    commit("plugin/shell: implement first-party sandboxed shell execution plugin", next(dates_iter), [
        "pkg/plugin/builtin/shell/shell_plugin.go"
    ])

    # Commit 5.4: Event bus, pub-sub routing, and event serialization
    write_file("pkg/events/bus.go", """package events

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
""")

    write_file("pkg/events/serialization.go", """package events

import (
	"encoding/json"
	"fmt"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func SerializeEvent(e *core.Event) ([]byte, error) {
	if e == nil {
		return nil, fmt.Errorf("cannot serialize nil event")
	}
	return json.Marshal(e)
}

func DeserializeEvent(data []byte) (*core.Event, error) {
	var e core.Event
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("failed to deserialize event: %w", err)
	}
	return &e, nil
}
""")

    # Commit 5.5: Webhook dispatcher with HMAC-SHA256 signatures and dead-letter handling
    write_file("pkg/webhook/dispatcher.go", """package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/retry"
)

type Target struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	SecretKey string `json:"secret_key"`
}

type DeadLetter struct {
	ID           core.ID     `json:"id"`
	TargetID     string      `json:"target_id"`
	Event        *core.Event `json:"event"`
	ErrorMessage string      `json:"error_message"`
	FailedAt     time.Time   `json:"failed_at"`
}

type Dispatcher struct {
	client      *http.Client
	policy      *retry.Policy
	deadLetterMu sync.RWMutex
	deadLetters  []*DeadLetter
}

func NewDispatcher(client *http.Client) *Dispatcher {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	backoff := retry.NewExponentialBackoff(100*time.Millisecond, time.Second, 2.0, false)
	return &Dispatcher{
		client:      client,
		policy:      retry.NewPolicy(3, backoff),
		deadLetters: make([]*DeadLetter, 0),
	}
}

func (d *Dispatcher) ComputeSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func (d *Dispatcher) Dispatch(ctx context.Context, target Target, event *core.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	sig := d.ComputeSignature(payload, target.SecretKey)

	err = d.policy.Execute(ctx, func(attemptCtx context.Context, attempt int) error {
		req, err := http.NewRequestWithContext(attemptCtx, http.MethodPost, target.URL, bytes.NewBuffer(payload))
		if err != nil {
			return retry.MarkNonRetryable(err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Kestrel-Event", string(event.Type))
		req.Header.Set("X-Kestrel-Signature", "sha256="+sig)

		resp, err := d.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return retry.MarkNonRetryable(fmt.Errorf("webhook client error: %d", resp.StatusCode))
		}
		return fmt.Errorf("webhook server error: %d", resp.StatusCode)
	})

	if err != nil {
		d.deadLetterMu.Lock()
		d.deadLetters = append(d.deadLetters, &DeadLetter{
			ID:           core.NewID("dl"),
			TargetID:     target.ID,
			Event:        event,
			ErrorMessage: err.Error(),
			FailedAt:     time.Now().UTC(),
		})
		d.deadLetterMu.Unlock()
	}

	return err
}

func (d *Dispatcher) DeadLetters() []*DeadLetter {
	d.deadLetterMu.RLock()
	defer d.deadLetterMu.RUnlock()
	out := make([]*DeadLetter, len(d.deadLetters))
	copy(out, d.deadLetters)
	return out
}
""")
    commit("events/webhook: implement event bus and webhook dispatcher with HMAC signatures", next(dates_iter), [
        "pkg/events/bus.go",
        "pkg/events/serialization.go",
        "pkg/webhook/dispatcher.go"
    ])

    # Commit 5.6: Round-trip serialization tests for EVERY event type and webhook test
    write_file("pkg/events/serialization_test.go", """package events

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestEventRoundTripSerialization(t *testing.T) {
	// Must test every domain event type for round-trip fidelity
	allEventTypes := []core.EventType{
		core.EventWorkflowCreated,
		core.EventWorkflowUpdated,
		core.EventWorkflowDeleted,
		core.EventRunStarted,
		core.EventRunCompleted,
		core.EventRunFailed,
		core.EventRunCancelled,
		core.EventRunSuspended,
		core.EventRunResumed,
		core.EventStepScheduled,
		core.EventStepStarted,
		core.EventStepCompleted,
		core.EventStepFailed,
		core.EventStepRetrying,
		core.EventStepSkipped,
		core.EventStepCancelled,
	}

	now := time.Now().UTC().Truncate(time.Millisecond)

	for _, et := range allEventTypes {
		t.Run(string(et), func(t *testing.T) {
			original := &core.Event{
				ID:        core.NewID("event"),
				Type:      et,
				TenantID:  "tenant-serialization-test",
				RunID:     core.NewID("run"),
				StepID:    "step-roundtrip",
				Timestamp: now,
				Payload:   json.RawMessage(`{"event_meta":"ok","type":"` + string(et) + `"}`),
			}

			// Encode
			encoded, err := SerializeEvent(original)
			if err != nil {
				t.Fatalf("failed serializing event %s: %v", et, err)
			}

			// Decode
			decoded, err := DeserializeEvent(encoded)
			if err != nil {
				t.Fatalf("failed deserializing event %s: %v", et, err)
			}

			// Deep comparison
			if decoded.ID != original.ID {
				t.Errorf("ID mismatch for %s: got %s, expected %s", et, decoded.ID, original.ID)
			}
			if decoded.Type != original.Type {
				t.Errorf("Type mismatch: got %s, expected %s", decoded.Type, original.Type)
			}
			if decoded.TenantID != original.TenantID {
				t.Errorf("TenantID mismatch: got %s, expected %s", decoded.TenantID, original.TenantID)
			}
			if decoded.RunID != original.RunID {
				t.Errorf("RunID mismatch: got %s, expected %s", decoded.RunID, original.RunID)
			}
			if decoded.StepID != original.StepID {
				t.Errorf("StepID mismatch: got %s, expected %s", decoded.StepID, original.StepID)
			}
			if !decoded.Timestamp.Equal(original.Timestamp) {
				t.Errorf("Timestamp mismatch: got %v, expected %v", decoded.Timestamp, original.Timestamp)
			}
			if !reflect.DeepEqual(decoded.Payload, original.Payload) {
				t.Errorf("Payload mismatch: got %s, expected %s", string(decoded.Payload), string(original.Payload))
			}
		})
	}
}
""")

    write_file("pkg/webhook/webhook_test.go", """package webhook

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
""")
    commit("events: add round-trip serialization tests for every event type and webhook suite", next(dates_iter), [
        "pkg/events/serialization_test.go",
        "pkg/webhook/webhook_test.go"
    ])

    print("Phase 5 completed successfully.")

if __name__ == '__main__':
    from generator.dates import generate_commit_dates
    from generator.git_utils import get_commit_count
    dates = iter(generate_commit_dates(180))
    for _ in range(get_commit_count()):
        next(dates)
    run_phase_5(dates)
