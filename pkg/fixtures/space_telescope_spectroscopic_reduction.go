package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// SpaceTelescopeSpectroscopicReductionPipeline builds a space observatory infrared spectral reduction DAG.
func SpaceTelescopeSpectroscopicReductionPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_jwst_nirspec_reduction"),
		TenantID:    "astrophysics-data-center",
		Name:        "JWST NIRSpec Multi-Object Near-Infrared Spectroscopic Data Reduction DAG",
		Version:     1,
		Description: "Reduces raw up-the-ramp detector pixel ramps, subtracts dark currents and snowballs, flags cosmic rays with Laplacian filter, extracts 2D micro-shutter array (MSA) slitlets into calibrated 1D flux spectra, and measures redshift absorption lines for high-z galaxies.",
		Timeout:     120 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "detector-ramp-fit-and-jump-detect",
				TaskType: "detector_ramp_fitting",
			},
			{
				ID:        "dark-current-and-bias-subtraction",
				TaskType:  "dark_bias_subtraction",
				DependsOn: []string{"detector-ramp-fit-and-jump-detect"},
			},
			{
				ID:        "msa-slitlet-2d-extraction",
				TaskType:  "msa_slitlet_extraction",
				DependsOn: []string{"dark-current-and-bias-subtraction"},
			},
			{
				ID:        "spectrophotometric-flux-calibration",
				TaskType:  "flux_calibration_1d",
				DependsOn: []string{"msa-slitlet-2d-extraction"},
			},
		},
	}
}

// DetectorRampFittingHandler fits linear regression across multi-accum frames and flags cosmic ray jumps.
type DetectorRampFittingHandler struct{}

func (h *DetectorRampFittingHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"num_integrations":10,"num_groups":20,"cosmic_ray_jumps_detected":412,"gain_electrons_per_dn":1.12}`),
	}, nil
}

// DarkBiasSubtractionHandler subtracts reference pixel bias and master super-dark frames.
type DarkBiasSubtractionHandler struct{}

func (h *DarkBiasSubtractionHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"reference_pixel_drift_dn":0.42,"mean_dark_rate_e_per_s":0.012,"snowball_cores_masked":18}`),
	}, nil
}

// MSASlitletExtractionHandler extracts curved dispersed spectral traces through micro-shutter slits.
type MSASlitletExtractionHandler struct{}

func (h *MSASlitletExtractionHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"num_slitlets_extracted":248,"spectral_resolution_r":2700,"wavelength_min_microns":0.97,"wavelength_max_microns":5.27}`),
	}, nil
}

// FluxCalibration1DHandler produces absolute flux-calibrated 1D spectra and redshift z estimation.
type FluxCalibration1DHandler struct{}

func (h *FluxCalibration1DHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"spectroscopic_redshift_z":8.496,"oiii_emission_snr":24.5,"c_iv_equivalent_width_angstrom":12.8,"calibration_status":"CERTIFIED"}`),
	}, nil
}
