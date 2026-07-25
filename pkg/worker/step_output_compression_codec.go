package worker

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"sync"
)

var (
	ErrDecompressionFault = errors.New("failed to decompress payload")
)

// GzipPayloadCompressionCodec handles transparent gzip compression for large step intermediate outputs.
type GzipPayloadCompressionCodec struct {
	mu             sync.RWMutex
	minCompressLen int
}

// NewGzipPayloadCompressionCodec creates a compression utility with a size threshold.
func NewGzipPayloadCompressionCodec(minBytesToCompress int) *GzipPayloadCompressionCodec {
	if minBytesToCompress <= 0 {
		minBytesToCompress = 512
	}
	return &GzipPayloadCompressionCodec{
		minCompressLen: minBytesToCompress,
	}
}

// ShouldCompress checks if length exceeds compression threshold.
func (c *GzipPayloadCompressionCodec) ShouldCompress(payload []byte) bool {
	return len(payload) >= c.minCompressLen
}

// Compress returns gzip compressed bytes.
func (c *GzipPayloadCompressionCodec) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	if _, err := zw.Write(data); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decompress decodes gzip compressed payload.
func (c *GzipPayloadCompressionCodec) Decompress(compressed []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, errors.Join(ErrDecompressionFault, err)
	}
	defer zr.Close()

	out, err := io.ReadAll(zr)
	if err != nil {
		return nil, errors.Join(ErrDecompressionFault, err)
	}
	return out, nil
}
