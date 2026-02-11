package worker

import (
	"bytes"
	"testing"
)

func TestPayloadChunkingStreamer(t *testing.T) {
	streamer := NewPayloadChunkingStreamer(10) // 10 bytes per chunk

	original := []byte("The quick brown fox jumps over the lazy dog.")
	chunks := streamer.Split(original)

	if len(chunks) != 5 { // 44 bytes / 10 = 5 chunks
		t.Fatalf("expected 5 chunks, got %d", len(chunks))
	}

	reassembled, err := streamer.Reassemble(chunks)
	if err != nil {
		t.Fatalf("unexpected reassemble error: %v", err)
	}

	if !bytes.Equal(original, reassembled) {
		t.Errorf("reassembled data mismatch: expected '%s', got '%s'", original, reassembled)
	}

	// Corrupt checksum
	chunks[2].Checksum = "invalid-sha256"
	_, err = streamer.Reassemble(chunks)
	if err == nil {
		t.Error("expected checksum mismatch error, got nil")
	}
}
