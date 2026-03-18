package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// RadiologyAIImagingPipeline builds a clinical DICOM volumetric PACS CAD pipeline.
func RadiologyAIImagingPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_radiology_ai_cad"),
		TenantID:    "clinical-radiology-network",
		Name:        "Chest CT Volumetric Nodule CAD & RECIST 1.1 Measurement Pipeline",
		Version:     1,
		Description: "Receives multi-slice DICOM studies via DICOMweb C-STORE, reconstructs isotropic 1mm voxel volumes, runs 3D U-Net pulmonary nodule segmentation, measures RECIST 1.1 longest diameters, and outputs structured FHIR DiagnosticReport with DICOM-SR measurements.",
		Timeout:     150 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "dicomweb-cstore-ingest",
				TaskType: "dicom_cstore_ingest",
			},
			{
				ID:        "isotropic-voxel-reconstruction",
				TaskType:  "voxel_reconstruction",
				DependsOn: []string{"dicomweb-cstore-ingest"},
			},
			{
				ID:        "unet3d-pulmonary-segmentation",
				TaskType:  "unet3d_segmentation",
				DependsOn: []string{"isotropic-voxel-reconstruction"},
			},
			{
				ID:        "recist-nodule-measurement",
				TaskType:  "recist_measurement",
				DependsOn: []string{"unet3d-pulmonary-segmentation"},
			},
			{
				ID:        "fhir-dicom-sr-report-gen",
				TaskType:  "fhir_dicomsr_gen",
				DependsOn: []string{"recist-nodule-measurement"},
			},
		},
	}
}

// DICOMCStoreIngestHandler parses DICOM Part 10 binary tags and de-identifies PHI.
type DICOMCStoreIngestHandler struct{}

func (h *DICOMCStoreIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"study_instance_uid":"1.2.840.113619.2.55","slices_count":512,"series_description":"THORAX_1MM_LUNG","kvp":120,"tube_current_ma":240}`),
	}, nil
}

// VoxelReconstructionHandler resamples Hounsfield Units into isotropic 1.0mm grids.
type VoxelReconstructionHandler struct{}

func (h *VoxelReconstructionHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"volume_dimensions":[512,512,512],"hu_range":[-1024,1450],"lung_window_level":-600,"lung_window_width":1500}`),
	}, nil
}

// UNet3DSegmentationHandler executes volumetric convolutional neural network inference.
type UNet3DSegmentationHandler struct{}

func (h *UNet3DSegmentationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"nodules_detected":2,"inference_time_seconds":8.4,"model_checkpoint":"unet3d_lung_v4","dice_score":0.912}`),
	}, nil
}

// RECISTMeasurementHandler calculates 3D longest axis and volume in cubic millimeters.
type RECISTMeasurementHandler struct{}

func (h *RECISTMeasurementHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"target_lesion_1_longest_diameter_mm":8.4,"target_lesion_1_volume_mm3":312.5,"spiculation_detected":true,"lung_rads_score":"4A"}`),
	}, nil
}

// FHIRDICOMSRGenHandler encodes findings into HL7 FHIR DiagnosticReport and DICOM-SR TID 1500.
type FHIRDICOMSRGenHandler struct{}

func (h *FHIRDICOMSRGenHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"fhir_diagnostic_report_id":"DX-2026-CT-0941","dicom_sr_sop_instance_uid":"1.2.840.113619.2.55.99","pacs_status":"COMMITTED"}`),
	}, nil
}
