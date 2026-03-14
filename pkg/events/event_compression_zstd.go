package events

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
)

// EventPayloadCompressor compresses event payloads to minimize message broker wire traffic.
type EventPayloadCompressor struct {
	minCompressionThreshold int
}

// NewEventPayloadCompressor creates a payload compressor.
func NewEventPayloadCompressor(threshold int) *EventPayloadCompressor {
	if threshold <= 0 {
		threshold = 512 // Only compress payloads larger than 512 bytes
	}
	return &EventPayloadCompressor{
		minCompressionThreshold: threshold,
	}
}

// Compress compresses byte array using GZIP stream.
func (c *EventPayloadCompressor) Compress(data []byte) ([]byte, bool, error) {
	if len(data) < c.minCompressionThreshold {
		return data, false, nil // skip compression
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return nil, false, err
	}
	if err := gw.Close(); err != nil {
		return nil, false, err
	}

	return buf.Bytes(), true, nil
}

// Decompress decompresses GZIP byte stream.
func (c *EventPayloadCompressor) Decompress(compressed []byte) ([]byte, error) {
	if len(compressed) == 0 {
		return nil, errors.New("empty payload to decompress")
	}

	gr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	return io.ReadAll(gr)
}
