package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// AutonomousDroneSurveyPipeline builds an industrial drone infrastructure inspection DAG.
func AutonomousDroneSurveyPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_drone_survey"),
		TenantID:    "geo-spatial-drones",
		Name:        "Autonomous Drone Inspection & Photogrammetry Pipeline",
		Version:     1,
		Description: "UAV RTK telemetry ingest, FAA LAANC airspace compliance, battery return-to-home margin check, Structure-from-Motion (SfM) point cloud assembly, thermal hotspot detection, 3D Poisson mesh reconstruction, and engineering report dispatch",
		Timeout:     120 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "telemetry-lidar-ingest",
				TaskType: "drone_ingest",
			},
			{
				ID:        "airspace-notam-validator",
				TaskType:  "airspace_check",
				DependsOn: []string{"telemetry-lidar-ingest"},
			},
			{
				ID:        "battery-geofence-safety-check",
				TaskType:  "safety_margins",
				DependsOn: []string{"airspace-notam-validator"},
			},
			{
				ID:        "point-cloud-orthomosaic-assembler",
				TaskType:  "sfm_reconstruct",
				DependsOn: []string{"battery-geofence-safety-check"},
			},
			{
				ID:        "thermal-defect-detector",
				TaskType:  "thermal_detect",
				DependsOn: []string{"point-cloud-orthomosaic-assembler"},
			},
			{
				ID:        "mesh-surface-reconstructor",
				TaskType:  "mesh_surface",
				DependsOn: []string{"thermal-defect-detector"},
			},
			{
				ID:        "engineering-inspection-report-generator",
				TaskType:  "report_generate",
				DependsOn: []string{"mesh-surface-reconstructor"},
			},
		},
	}
}

// DroneIngestHandler parses UAV flight logs, RTK GPS positions, and raw LiDAR packets.
type DroneIngestHandler struct{}

func (h *DroneIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"uav_id":"DJI-M300-RTK-09","flight_time_minutes":28,"lidar_points_captured":14500000,"photos_captured":450}`),
	}, nil
}

// AirspaceCheckHandler validates flight boundary against FAA Part 107 and temporary flight restrictions (TFRs).
type AirspaceCheckHandler struct{}

func (h *AirspaceCheckHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"airspace_class":"Class-G","laanc_authorization_id":"FAA-AUTH-9921","tfr_active":false,"compliant":true}`),
	}, nil
}

// SafetyMarginsHandler calculates return-to-home (RTH) wind vector resistance and battery reserve.
type SafetyMarginsHandler struct{}

func (h *SafetyMarginsHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"battery_remaining_pct":38,"rth_required_pct":18,"safety_margin_pct":20,"safe_landing_confirmed":true}`),
	}, nil
}

// SfMReconstructHandler runs Structure-from-Motion photogrammetry to align keypoints into an orthomosaic.
type SfMReconstructHandler struct{}

func (h *SfMReconstructHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"ground_sampling_distance_cm":1.2,"dense_cloud_points":42000000,"reprojection_error_px":0.48}`),
	}, nil
}

// ThermalDetectHandler flags solar panel bypass diode failures or concrete delaminations via FLIR radiometry.
type ThermalDetectHandler struct{}

func (h *ThermalDetectHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"thermal_anomalies_detected":2,"max_temp_delta_c":14.5,"criticality":"MODERATE","location":"Array-B-String-4"}`),
	}, nil
}

// MeshSurfaceHandler builds Poisson water-tight polygon surfaces from point cloud coordinates.
type MeshSurfaceHandler struct{}

func (h *MeshSurfaceHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"mesh_format":"OBJ","triangle_count":8500000,"texture_resolution":"8K","model_file":"s3://survey-models/mission-441.obj"}`),
	}, nil
}

// DroneReportGenerateHandler compiles the PE (Professional Engineer) stamped inspection review packet.
type DroneReportGenerateHandler struct{}

func (h *DroneReportGenerateHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	reportID := fmt.Sprintf("REP-UAV-%d", time.Now().UnixNano())
	payload := fmt.Sprintf(`{"report_id":"%s","status":"READY_FOR_ENGINEER_STAMP","pdf_url":"https://inspections.net/r/%s.pdf"}`, reportID, reportID)
	return &worker.StepResult{Output: json.RawMessage(payload)}, nil
}
