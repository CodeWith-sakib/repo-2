package events

import (
	"testing"
)

func TestEventFilterGroup(t *testing.T) {
	rules := []EventPropertyRule{
		{
			Field: "tenant.id",
			Op:    OpEquals,
			Value: "acme-corp",
		},
		{
			Field: "event_type",
			Op:    OpPrefix,
			Value: "order.",
		},
		{
			Field: "payload.amount",
			Op:    OpExists,
		},
		{
			Field: "trace_id",
			Op:    OpRegexMatch,
			Value: `^trc-[0-9]{4}$`,
		},
	}

	group, err := NewEventFilterGroup(rules)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	matchingPayload := []byte(`{
		"tenant": {"id": "acme-corp"},
		"event_type": "order.completed",
		"payload": {"amount": 250.0},
		"trace_id": "trc-4091"
	}`)

	if !group.Matches(matchingPayload) {
		t.Errorf("expected matching payload to pass filter")
	}

	nonMatchingPayload := []byte(`{
		"tenant": {"id": "other-corp"},
		"event_type": "order.completed",
		"payload": {"amount": 250.0},
		"trace_id": "trc-4091"
	}`)

	if group.Matches(nonMatchingPayload) {
		t.Errorf("expected non-matching payload to fail filter")
	}
}
