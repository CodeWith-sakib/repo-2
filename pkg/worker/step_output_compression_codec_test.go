package worker

import (
	"bytes"
	"strings"
	"testing"
)

func TestGzipPayloadCompressionCodec(t *testing.T) {
	codec := NewGzipPayloadCompressionCodec(100)

	raw := []byte(strings.Repeat("structured-json-telemetry-block-with-repeating-keys-", 20))
	if !codec.ShouldCompress(raw) {
		t.Fatal("large payload should trigger compression")
	}

	compressed, err := codec.Compress(raw)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	if len(compressed) >= len(raw) {
		t.Errorf("compression did not reduce size: %d vs %d", len(compressed), len(raw))
	}

	decompressed, err := codec.Decompress(compressed)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !bytes.Equal(decompressed, raw) {
		t.Error("decompressed data does not match original")
	}
}
