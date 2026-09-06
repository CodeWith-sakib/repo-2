package events

import (
	"encoding/json"
	"sync"
	"time"
)

// DeadLetterEventEnvelope encapsulates a poisoned or rejected event with forensic diagnosis.
type DeadLetterEventEnvelope struct {
	OriginalTopic    string            `json:"original_topic"`
	OriginalPayload  string            `json:"original_payload"`
	FailureReason    string            `json:"failure_reason"`
	AttemptCount     int               `json:"attempt_count"`
	OriginatingNode  string            `json:"originating_node"`
	CapturedMetadata map[string]string `json:"captured_metadata"`
	PackagedAt       time.Time         `json:"packaged_at"`
}

// DeadLetterEnvelopePacker serializes forensic envelopes for DLQ persistence.
type DeadLetterEnvelopePacker struct {
	mu sync.RWMutex
}

// NewDeadLetterEnvelopePacker creates an envelope formatter.
func NewDeadLetterEnvelopePacker() *DeadLetterEnvelopePacker {
	return &DeadLetterEnvelopePacker{}
}

// Pack marshals the failure envelope into formatted JSON bytes.
func (p *DeadLetterEnvelopePacker) Pack(topic string, payload []byte, reason string, attempts int, node string, meta map[string]string) ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	envelope := DeadLetterEventEnvelope{
		OriginalTopic:    topic,
		OriginalPayload:  string(payload),
		FailureReason:    reason,
		AttemptCount:     attempts,
		OriginatingNode:  node,
		CapturedMetadata: meta,
		PackagedAt:       time.Now(),
	}

	return json.Marshal(envelope)
}

// Unpack deserializes a raw DLQ event payload into a structured envelope.
func (p *DeadLetterEnvelopePacker) Unpack(data []byte) (*DeadLetterEventEnvelope, error) {
	var env DeadLetterEventEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return &env, nil
}
