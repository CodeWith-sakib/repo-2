package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestRadiologyAIImagingPipeline_Validate(t *testing.T) {
	wf := RadiologyAIImagingPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestRadiologyAIImagingPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-rad-01", StepID: "dicomweb-cstore-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"dicom_cstore_ingest", &DICOMCStoreIngestHandler{}},
		{"voxel_reconstruction", &VoxelReconstructionHandler{}},
		{"unet3d_segmentation", &UNet3DSegmentationHandler{}},
		{"recist_measurement", &RECISTMeasurementHandler{}},
		{"fhir_dicomsr_gen", &FHIRDICOMSRGenHandler{}},
	}

	for _, h := range handlers {
		res, err := h.handler.Execute(ctx, sctx)
		if err != nil {
			t.Fatalf("handler %s failed: %v", h.name, err)
		}
		if len(res.Output) == 0 {
			t.Errorf("expected non-empty output for handler %s", h.name)
		}
	}
}
