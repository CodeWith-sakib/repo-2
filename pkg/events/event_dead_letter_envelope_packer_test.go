package events

import (
	"testing"
)

func TestDeadLetterEnvelopePacker(t *testing.T) {
	packer := NewDeadLetterEnvelopePacker()

	bytes, err := packer.Pack("telemetry.orders", []byte(`{"id": 123}`), "JSON validation error", 3, "worker-node-1", map[string]string{"region": "us-west-2"})
	if err != nil {
		t.Fatalf("packing failed: %v", err)
	}

	unpacked, err := packer.Unpack(bytes)
	if err != nil {
		t.Fatalf("unpacking failed: %v", err)
	}
	if unpacked.OriginalTopic != "telemetry.orders" {
		t.Errorf("topic mismatch: %s", unpacked.OriginalTopic)
	}
	if unpacked.AttemptCount != 3 {
		t.Errorf("attempts mismatch: %d", unpacked.AttemptCount)
	}
}
