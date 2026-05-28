package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// GravitationalWaveLIGOCoincidencePipeline builds a dual-interferometer gravitational wave transient detection DAG.
func GravitationalWaveLIGOCoincidencePipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_gw_ligo_coincidence"),
		TenantID:    "gravitational-physics-observatory",
		Name:        "LIGO-Virgo-KAGRA 4kHz Strain Time-Series Gravitational Wave Coincidence DAG",
		Version:     1,
		Description: "Ingests calibrated 4096Hz h(t) strain feeds from Hanford and Livingston observatories, executes Q-transform multi-resolution glitch whitening, correlates matched filtering across 250,000 binary black hole (BBH) post-Newtonian waveform templates, checks 10ms light travel time coincidence, and emits GCN notice for multi-messenger optical follow-up.",
		Timeout:     60 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "hanford-livingston-strain-ingest",
				TaskType: "strain_series_ingest",
			},
			{
				ID:        "qtransform-spectral-whitening",
				TaskType:  "qtransform_whitening",
				DependsOn: []string{"hanford-livingston-strain-ingest"},
			},
			{
				ID:        "bbh-template-matched-filtering",
				TaskType:  "matched_filter_bank",
				DependsOn: []string{"qtransform-spectral-whitening"},
			},
			{
				ID:        "interferometer-temporal-coincidence",
				TaskType:  "temporal_coincidence_eval",
				DependsOn: []string{"bbh-template-matched-filtering"},
			},
			{
				ID:        "gcn-astronomy-skymap-broadcast",
				TaskType:  "gcn_skymap_broadcast",
				DependsOn: []string{"interferometer-temporal-coincidence"},
			},
		},
	}
}

// StrainSeriesIngestHandler validates calibration state vectors and power spectral density (PSD).
type StrainSeriesIngestHandler struct{}

func (h *StrainSeriesIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"sampling_rate_hz":4096,"observatories":["LIGO_HANFORD","LIGO_LIVINGSTON"],"laser_power_watts":45.0,"bns_range_mpc":142.5}`),
	}, nil
}

// QTransformWhiteningHandler applies overcomplete sine-Gaussian tile transforms to whiten colored noise.
type QTransformWhiteningHandler struct{}

func (h *QTransformWhiteningHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"peak_energy":42.5,"central_frequency_hz":128.4,"glitch_veto_passed":true,"snr_loss_pct":0.8}`),
	}, nil
}

// MatchedFilterBankHandler convolves Fourier strain against numerical relativity waveforms.
type MatchedFilterBankHandler struct{}

func (h *MatchedFilterBankHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"primary_mass_solar":36.2,"secondary_mass_solar":29.1,"chirp_mass_solar":28.4,"effective_spin_chi_eff":0.04,"network_snr":24.5}`),
	}, nil
}

// TemporalCoincidenceEvalHandler enforces light travel time delta between sites (|dt| <= 10.0ms).
type TemporalCoincidenceEvalHandler struct{}

func (h *TemporalCoincidenceEvalHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"hanford_livingston_dt_ms":6.8,"false_alarm_rate_years":1420000,"significance":"5_SIGMA_DISCOVERY"}`),
	}, nil
}

// GCNSkymapBroadcastHandler generates HEALPix probability localization map and triggers Gamma-ray optical telescopes.
type GCNSkymapBroadcastHandler struct{}

func (h *GCNSkymapBroadcastHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"event_designation":"GW20260603A","sky_area_90_deg2":38.4,"luminosity_distance_mpc":410.0,"gcn_circular_id":38942,"status":"BROADCASTED_TO_ASTRONOMERS"}`),
	}, nil
}
