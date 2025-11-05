package worker

import (
	"bytes"
	"strings"
	"testing"
)

func TestPayloadCodecManager_GzipRoundTrip(t *testing.T) {
	mgr := NewPayloadCodecManager(50) // threshold = 50 bytes

	// Large repetitive payload -> should trigger gzip
	raw := []byte(strings.Repeat("Workflow execution state tracking payload\n", 20))

	encoded, err := mgr.Encode(raw)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	if encoded.Codec != CodecGzip {
		t.Errorf("expected GZIP codec, got %s", encoded.Codec)
	}
	if encoded.EncodedBytesCount >= encoded.RawBytesCount {
		t.Errorf("expected compression savings: compressed=%d, raw=%d", encoded.EncodedBytesCount, encoded.RawBytesCount)
	}

	decoded, err := mgr.Decode(encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if !bytes.Equal(decoded, raw) {
		t.Error("decoded bytes do not match original")
	}
}

func TestPayloadCodecManager_UnderThreshold(t *testing.T) {
	mgr := NewPayloadCodecManager(100)

	small := []byte("short payload")
	encoded, err := mgr.Encode(small)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	if encoded.Codec != CodecNone {
		t.Errorf("expected CodecNone for small payload, got %s", encoded.Codec)
	}

	decoded, err := mgr.Decode(encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if !bytes.Equal(decoded, small) {
		t.Error("decoded bytes mismatch")
	}
}
