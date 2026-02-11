package worker

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// PayloadChunk represents a single sliced packet of a large workflow payload.
type PayloadChunk struct {
	ChunkIndex int    `json:"chunk_index"`
	TotalChunks int   `json:"total_chunks"`
	Data       []byte `json:"data"`
	Checksum   string `json:"checksum"`
}

// PayloadChunkingStreamer slices large inputs/outputs into bounded transmission blocks.
type PayloadChunkingStreamer struct {
	chunkSize int
}

// NewPayloadChunkingStreamer creates a payload chunker with specified max chunk size in bytes.
func NewPayloadChunkingStreamer(chunkSize int) *PayloadChunkingStreamer {
	if chunkSize <= 0 {
		chunkSize = 64 * 1024 // 64KB default
	}
	return &PayloadChunkingStreamer{
		chunkSize: chunkSize,
	}
}

// Split divides data buffer into ordered PayloadChunks.
func (s *PayloadChunkingStreamer) Split(data []byte) []PayloadChunk {
	if len(data) == 0 {
		return nil
	}

	total := (len(data) + s.chunkSize - 1) / s.chunkSize
	chunks := make([]PayloadChunk, total)

	for i := 0; i < total; i++ {
		start := i * s.chunkSize
		end := start + s.chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunkData := data[start:end]
		h := sha256.Sum256(chunkData)

		chunks[i] = PayloadChunk{
			ChunkIndex:  i,
			TotalChunks: total,
			Data:        chunkData,
			Checksum:    hex.EncodeToString(h[:]),
		}
	}

	return chunks
}

// Reassemble verifies chunk checksums and reconstructs the full original buffer.
func (s *PayloadChunkingStreamer) Reassemble(chunks []PayloadChunk) ([]byte, error) {
	if len(chunks) == 0 {
		return nil, nil
	}

	var buf bytes.Buffer
	for i, c := range chunks {
		if c.ChunkIndex != i {
			return nil, fmt.Errorf("chunk out of order: expected %d, got %d", i, c.ChunkIndex)
		}

		h := sha256.Sum256(c.Data)
		calcSum := hex.EncodeToString(h[:])
		if calcSum != c.Checksum {
			return nil, fmt.Errorf("chunk %d checksum mismatch: expected %s, calculated %s", i, c.Checksum, calcSum)
		}

		_, _ = buf.Write(c.Data)
	}

	return buf.Bytes(), nil
}

// StreamToWriter streams all chunks sequentially into an io.Writer.
func (s *PayloadChunkingStreamer) StreamToWriter(w io.Writer, chunks []PayloadChunk) error {
	data, err := s.Reassemble(chunks)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
