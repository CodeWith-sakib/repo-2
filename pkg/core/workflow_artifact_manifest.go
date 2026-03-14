package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// WorkflowArtifactDescriptor stores metadata and hash digest of a produced artifact.
type WorkflowArtifactDescriptor struct {
	ArtifactID   string    `json:"artifact_id"`
	StepID       string    `json:"step_id"`
	Name         string    `json:"name"`
	URI          string    `json:"uri"`
	SizeBytes    int64     `json:"size_bytes"`
	SHA256Hex    string    `json:"sha256_hex"`
	MimeType     string    `json:"mime_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// WorkflowArtifactManifest tracks all intermediate and terminal artifacts generated during execution.
type WorkflowArtifactManifest struct {
	mu        sync.RWMutex
	runID     string
	artifacts map[string]WorkflowArtifactDescriptor
}

// NewWorkflowArtifactManifest creates an artifact manifest for a workflow run.
func NewWorkflowArtifactManifest(runID string) *WorkflowArtifactManifest {
	return &WorkflowArtifactManifest{
		runID:     runID,
		artifacts: make(map[string]WorkflowArtifactDescriptor),
	}
}

// RegisterArtifact records an artifact with checksum calculation.
func (m *WorkflowArtifactManifest) RegisterArtifact(stepID, name, uri string, data []byte, mimeType string) (WorkflowArtifactDescriptor, error) {
	if stepID == "" || name == "" || uri == "" {
		return WorkflowArtifactDescriptor{}, fmt.Errorf("stepID, name, and uri are required")
	}

	h := sha256.Sum256(data)
	hashHex := hex.EncodeToString(h[:])

	desc := WorkflowArtifactDescriptor{
		ArtifactID: fmt.Sprintf("art_%s_%s", stepID, name),
		StepID:     stepID,
		Name:       name,
		URI:        uri,
		SizeBytes:  int64(len(data)),
		SHA256Hex:  hashHex,
		MimeType:   mimeType,
		CreatedAt:  time.Now().UTC(),
	}

	m.mu.Lock()
	m.artifacts[desc.ArtifactID] = desc
	m.mu.Unlock()

	return desc, nil
}

// GetArtifact retrieves artifact metadata by its unique descriptor ID.
func (m *WorkflowArtifactManifest) GetArtifact(artifactID string) (WorkflowArtifactDescriptor, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	desc, exists := m.artifacts[artifactID]
	return desc, exists
}

// ListArtifacts returns all tracked artifacts for the workflow run.
func (m *WorkflowArtifactManifest) ListArtifacts() []WorkflowArtifactDescriptor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]WorkflowArtifactDescriptor, 0, len(m.artifacts))
	for _, a := range m.artifacts {
		list = append(list, a)
	}
	return list
}
