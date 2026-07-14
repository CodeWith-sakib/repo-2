package core

import (
	"testing"
)

func TestProvenanceManifestBuilder(t *testing.T) {
	builder := NewProvenanceManifestBuilder("wf-prov-1", "run-100")

	payload1 := []byte("model-weights-binary-data")
	art1 := builder.RecordArtifact("step-train", "s3://models/weights.bin", payload1, "application/octet-stream", map[string]string{"epoch": "50"})
	if art1.SizeBytes != int64(len(payload1)) {
		t.Errorf("expected size %d, got %d", len(payload1), art1.SizeBytes)
	}
	if art1.SHA256Hash == "" {
		t.Error("expected non-empty hash")
	}

	manifest := builder.Seal()
	if manifest.WorkflowID != "wf-prov-1" || manifest.RunID != "run-100" {
		t.Errorf("manifest IDs mismatch")
	}
	if len(manifest.Artifacts) != 1 {
		t.Errorf("expected 1 artifact, got %d", len(manifest.Artifacts))
	}
	if manifest.TotalBytes != int64(len(payload1)) {
		t.Errorf("expected total bytes %d, got %d", len(payload1), manifest.TotalBytes)
	}
}
