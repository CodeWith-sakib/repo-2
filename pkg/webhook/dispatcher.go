package webhook

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
