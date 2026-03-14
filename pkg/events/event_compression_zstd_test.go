package events

import (
	"bytes"
	"strings"
	"testing"
)

func TestEventPayloadCompressor(t *testing.T) {
	compressor := NewEventPayloadCompressor(100)

	// Below threshold -> uncompressed pass-through
	small := []byte("small payload")
	out, wasCompressed, err := compressor.Compress(small)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wasCompressed {
		t.Error("expected small payload to not be compressed")
	}
	if !bytes.Equal(small, out) {
		t.Error("expected unchanged small payload")
	}

	// Above threshold -> compressed
	large := []byte(strings.Repeat("event_data_stream_json_payload_field_", 20))
	comp, wasCompressed, err := compressor.Compress(large)
	if err != nil {
		t.Fatalf("unexpected error compressing large: %v", err)
	}
	if !wasCompressed {
		t.Error("expected large payload to be compressed")
	}
	if len(comp) >= len(large) {
		t.Errorf("expected compression reduction: comp=%d, orig=%d", len(comp), len(large))
	}

	// Decompress roundtrip
	decomp, err := compressor.Decompress(comp)
	if err != nil {
		t.Fatalf("unexpected decompression error: %v", err)
	}
	if !bytes.Equal(large, decomp) {
		t.Error("decompressed data mismatch")
	}
}
