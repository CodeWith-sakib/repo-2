package auth

import (
	"crypto/rand"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

type AuditAction string

const (
	AuditWorkflowCreated AuditAction = "workflow.created"
	AuditWorkflowUpdated AuditAction = "workflow.updated"
	AuditWorkflowDeleted AuditAction = "workflow.deleted"
	AuditRunTriggered   AuditAction = "run.triggered"
	AuditRunCancelled   AuditAction = "run.cancelled"
)

type AuditEvent struct {
	ID        string      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Actor     string      `json:"actor"`
	Action    AuditAction `json:"action"`
	TargetID  string      `json:"target_id"`
	Details   string      `json:"details,omitempty"`
	Signature string      `json:"signature"`
}

type AuditLogger struct {
	mu     sync.RWMutex
	secret []byte
	events []*AuditEvent
}

func NewAuditLogger(signingKey string) *AuditLogger {
	return &AuditLogger{
		secret: []byte(signingKey),
		events: make([]*AuditEvent, 0),
	}
}

func (l *AuditLogger) Log(actor string, action AuditAction, targetID, details string) (*AuditEvent, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	evt := &AuditEvent{
		ID:        randomHex(8),
		Timestamp: time.Now().UTC(),
		Actor:     actor,
		Action:    action,
		TargetID:  targetID,
		Details:   details,
	}

	sig, err := l.signEvent(evt)
	if err != nil {
		return nil, err
	}
	evt.Signature = sig

	l.events = append(l.events, evt)
	return evt, nil
}

func (l *AuditLogger) Verify(evt *AuditEvent) bool {
	sig, err := l.signEvent(evt)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(evt.Signature), []byte(sig))
}

func (l *AuditLogger) signEvent(evt *AuditEvent) (string, error) {
	mac := hmac.New(sha256.New, l.secret)
	data := evt.ID + "|" + evt.Timestamp.Format(time.RFC3339Nano) + "|" + evt.Actor + "|" + string(evt.Action) + "|" + evt.TargetID + "|" + evt.Details
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (l *AuditLogger) Events() []*AuditEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]*AuditEvent, len(l.events))
	copy(out, l.events)
	return out
}

func (evt *AuditEvent) ToJSON() ([]byte, error) {
	return json.Marshal(evt)
}


func randomHex(bytesLen int) string {
	b := make([]byte, bytesLen)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
