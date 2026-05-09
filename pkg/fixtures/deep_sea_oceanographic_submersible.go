package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// DeepSeaOceanographicSubmersiblePipeline builds an autonomous underwater vehicle (AUV) hydrothermal vent survey DAG.
func DeepSeaOceanographicSubmersiblePipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_auv_hadal_trench_survey"),
		TenantID:    "oceanographic-exploration-institute",
		Name:        "6000m Hadal Trench Autonomous Submersible Hydrothermal Vent Exploration DAG",
		Version:     1,
		Description: "Ingests multibeam bathymetric sonar point clouds at 6000m depth (600 bar hydrostatic pressure), executes Doppler Velocity Log (DVL) dead reckoning navigation, maps methane/temperature hydrothermal plume gradients, executes stereo photogrammetry 3D benthic reconstruction, and generates acoustic USBL emergency rendezvous surfacing vector.",
		Timeout:     120 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "multibeam-sonar-bathymetry-ingest",
				TaskType: "multibeam_sonar_ingest",
			},
			{
				ID:        "dvl-ins-dead-reckoning-nav",
				TaskType:  "dvl_inertial_navigation",
				DependsOn: []string{"multibeam-sonar-bathymetry-ingest"},
			},
			{
				ID:        "hydrothermal-plume-gradient-trace",
				TaskType:  "plume_gradient_trace",
				DependsOn: []string{"dvl-ins-dead-reckoning-nav"},
			},
			{
				ID:        "stereo-photogrammetry-benthic-map",
				TaskType:  "photogrammetry_benthic_map",
				DependsOn: []string{"hydrothermal-plume-gradient-trace"},
			},
			{
				ID:        "usbl-acoustic-ascent-vector-calc",
				TaskType:  "usbl_ascent_vector_calc",
				DependsOn: []string{"stereo-photogrammetry-benthic-map"},
			},
		},
	}
}

// MultibeamSonarIngestHandler cleans seafloor backscatter intensity and sound velocity profiles.
type MultibeamSonarIngestHandler struct{}

func (h *MultibeamSonarIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"depth_meters":6240.5,"hydrostatic_pressure_bar":624.1,"sound_velocity_mps":1522.4,"swath_width_meters":450}`),
	}, nil
}

// DVLInertialNavigationHandler integrates Doppler bottom-track velocity vectors and ring laser gyro heading.
type DVLInertialNavigationHandler struct{}

func (h *DVLInertialNavigationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"altitude_above_bottom_m":12.4,"speed_over_ground_knots":2.1,"drift_error_meters":0.42,"heading_deg":184.2}`),
	}, nil
}

// PlumeGradientTraceHandler traces oxidation-reduction potential (Eh) and dissolved methane (CH4) spikes.
type PlumeGradientTraceHandler struct{}

func (h *PlumeGradientTraceHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"chimney_detected":true,"max_vent_temp_c":342.5,"methane_nmol_l":8420.0,"plume_core_distance_m":8.5}`),
	}, nil
}

// PhotogrammetryBenthicMapHandler aligns overlapping 4K strobe images into mm-accurate seafloor orthomosaics.
type PhotogrammetryBenthicMapHandler struct{}

func (h *PhotogrammetryBenthicMapHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"images_aligned":1240,"orthomosaic_area_m2":1850.0,"vent_fauna_count":4820,"dominant_species":"Rimicaris_exoculata"}`),
	}, nil
}

// USBLAscentVectorHandler derives ballast jettison trajectory and acoustic transponder uplink.
type USBLAscentVectorHandler struct{}

func (h *USBLAscentVectorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"surface_eta_minutes":48.2,"ascent_rate_mps":2.15,"surface_ship_bearing_deg":42.0,"usbl_status":"ACOUSTIC_LOCK_ESTABLISHED"}`),
	}, nil
}
