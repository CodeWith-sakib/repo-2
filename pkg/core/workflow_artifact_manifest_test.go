package core

import (
	"testing"
)

func TestWorkflowArtifactManifest(t *testing.T) {
	manifest := NewWorkflowArtifactManifest("run-101")

	data := []byte("model weights v1.0 binary payload")
	desc, err := manifest.RegisterArtifact("train_step", "weights.bin", "s3://models/weights.bin", data, "application/octet-stream")
	if err != nil {
		t.Fatalf("unexpected error registering artifact: %v", err)
	}

	if desc.SizeBytes != int64(len(data)) {
		t.Errorf("expected size %d, got %d", len(data), desc.SizeBytes)
	}

	if len(desc.SHA256Hex) != 64 {
		t.Errorf("expected 64 char sha256 hex, got %s", desc.SHA256Hex)
	}

	retrieved, found := manifest.GetArtifact(desc.ArtifactID)
	if !found {
		t.Fatal("expected artifact to be retrieved")
	}

	if retrieved.Name != "weights.bin" {
		t.Errorf("expected name weights.bin, got %s", retrieved.Name)
	}

	all := manifest.ListArtifacts()
	if len(all) != 1 {
		t.Errorf("expected 1 artifact in list, got %d", len(all))
	}
}
