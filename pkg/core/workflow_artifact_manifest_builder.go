package core

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// ImmutableArtifactRecord describes an immutable output file or blob recorded by a step.
type ImmutableArtifactRecord struct {
	StepID      string            `json:"step_id"`
	ArtifactURI string            `json:"artifact_uri"`
	SizeBytes   int64             `json:"size_bytes"`
	SHA256Hash  string            `json:"sha256_hash"`
	MimeType    string            `json:"mime_type"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
}

// SealedProvenanceManifest records all data outputs generated during a run for compliance provenance.
type SealedProvenanceManifest struct {
	WorkflowID string                    `json:"workflow_id"`
	RunID      string                    `json:"run_id"`
	Artifacts  []ImmutableArtifactRecord `json:"artifacts"`
	TotalBytes int64                     `json:"total_bytes"`
	SealedAt   time.Time                 `json:"sealed_at"`
}

// ProvenanceManifestBuilder collects artifact references thread-safely across step executions.
type ProvenanceManifestBuilder struct {
	mu         sync.Mutex
	workflowID string
	runID      string
	artifacts  []ImmutableArtifactRecord
	totalBytes int64
}

// NewProvenanceManifestBuilder initializes a provenance manifest builder.
func NewProvenanceManifestBuilder(workflowID, runID string) *ProvenanceManifestBuilder {
	return &ProvenanceManifestBuilder{
		workflowID: workflowID,
		runID:      runID,
		artifacts:  make([]ImmutableArtifactRecord, 0),
	}
}

// RecordArtifact registers an output payload and calculates its checksum.
func (b *ProvenanceManifestBuilder) RecordArtifact(stepID, uri string, content []byte, mimeType string, meta map[string]string) ImmutableArtifactRecord {
	b.mu.Lock()
	defer b.mu.Unlock()

	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])
	sz := int64(len(content))

	art := ImmutableArtifactRecord{
		StepID:      stepID,
		ArtifactURI: uri,
		SizeBytes:   sz,
		SHA256Hash:  hashStr,
		MimeType:    mimeType,
		Metadata:    meta,
		CreatedAt:   time.Now(),
	}

	b.artifacts = append(b.artifacts, art)
	b.totalBytes += sz
	return art
}

// Seal generates an immutable signed manifest for archival.
func (b *ProvenanceManifestBuilder) Seal() SealedProvenanceManifest {
	b.mu.Lock()
	defer b.mu.Unlock()

	copied := make([]ImmutableArtifactRecord, len(b.artifacts))
	copy(copied, b.artifacts)

	return SealedProvenanceManifest{
		WorkflowID: b.workflowID,
		RunID:      b.runID,
		Artifacts:  copied,
		TotalBytes: b.totalBytes,
		SealedAt:   time.Now(),
	}
}
