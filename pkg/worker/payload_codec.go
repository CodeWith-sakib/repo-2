package worker

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
)

// CompressionCodec identifies compression type.
type CompressionCodec string

const (
	CodecNone CompressionCodec = "NONE"
	CodecGzip CompressionCodec = "GZIP"
)

// EncodedPayload stores serialized bytes with format metadata.
type EncodedPayload struct {
	Codec             CompressionCodec
	RawBytesCount     int
	EncodedBytesCount int
	Data              []byte
}

// PayloadCodecManager handles payload compression and decompression.
type PayloadCodecManager struct {
	thresholdBytes int
}

// NewPayloadCodecManager creates a manager with compression threshold.
func NewPayloadCodecManager(thresholdBytes int) *PayloadCodecManager {
	if thresholdBytes <= 0 {
		thresholdBytes = 1024 // 1 KB default
	}
	return &PayloadCodecManager{
		thresholdBytes: thresholdBytes,
	}
}

// Encode compresses data using Gzip if it exceeds thresholdBytes.
func (m *PayloadCodecManager) Encode(data []byte) (*EncodedPayload, error) {
	rawLen := len(data)
	if rawLen < m.thresholdBytes {
		return &EncodedPayload{
			Codec:             CodecNone,
			RawBytesCount:     rawLen,
			EncodedBytesCount: rawLen,
			Data:              data,
		}, nil
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return nil, fmt.Errorf("gzip compression failed: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("gzip close failed: %w", err)
	}

	compressed := buf.Bytes()
	return &EncodedPayload{
		Codec:             CodecGzip,
		RawBytesCount:     rawLen,
		EncodedBytesCount: len(compressed),
		Data:              compressed,
	}, nil
}

// Decode decompresses an EncodedPayload back to original bytes.
func (m *PayloadCodecManager) Decode(p *EncodedPayload) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("nil payload")
	}

	switch p.Codec {
	case CodecNone:
		return p.Data, nil

	case CodecGzip:
		gr, err := gzip.NewReader(bytes.NewReader(p.Data))
		if err != nil {
			return nil, fmt.Errorf("gzip reader failed: %w", err)
		}
		defer gr.Close()

		decompressed, err := io.ReadAll(gr)
		if err != nil {
			return nil, fmt.Errorf("gzip decompress failed: %w", err)
		}
		return decompressed, nil

	default:
		return nil, fmt.Errorf("unsupported codec: %s", p.Codec)
	}
}

// EncodeBase64 returns standard Base64 representation.
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 decodes standard Base64 string.
func DecodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
