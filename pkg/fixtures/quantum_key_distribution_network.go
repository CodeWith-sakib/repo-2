package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// QuantumKeyDistributionPipeline builds a BB84/E91 quantum satellite-to-ground optical entanglement DAG.
func QuantumKeyDistributionPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_qkd_optical_network"),
		TenantID:    "cybersecurity-quantum-crypt",
		Name:        "Satellite-to-Ground Entangled Photon BB84 QKD Key Exchange DAG",
		Version:     1,
		Description: "Establishes optical ground station beacon lock, transmits polarized single-photon pulses at 850nm, performs quantum bit error rate (QBER) estimation, executes cascade error correction, and derives one-time-pad AES-256-GCM symmetric keys via Toeplitz matrix privacy amplification.",
		Timeout:     60 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "optical-beacon-telescope-tracking",
				TaskType: "telescope_beacon_track",
			},
			{
				ID:        "single-photon-bb84-pulse-detect",
				TaskType:  "bb84_photon_detect",
				DependsOn: []string{"optical-beacon-telescope-tracking"},
			},
			{
				ID:        "qber-eavesdropping-sifting-check",
				TaskType:  "qber_sifting_check",
				DependsOn: []string{"single-photon-bb84-pulse-detect"},
			},
			{
				ID:        "cascade-shannon-error-correction",
				TaskType:  "cascade_error_correction",
				DependsOn: []string{"qber-eavesdropping-sifting-check"},
			},
			{
				ID:        "toeplitz-privacy-amplification",
				TaskType:  "toeplitz_privacy_amp",
				DependsOn: []string{"cascade-shannon-error-correction"},
			},
		},
	}
}

// TelescopeBeaconTrackHandler synchronizes laser acquisition and pointing tracking (PAT).
type TelescopeBeaconTrackHandler struct{}

func (h *TelescopeBeaconTrackHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"coarse_pointing_error_urad":1.4,"fine_piezo_tracking_lock":true,"optical_snr_db":34.5,"station":"TENERIFE_OGS"}`),
	}, nil
}

// BB84PhotonDetectHandler registers single-photon avalanche diode (SPAD) clicks across rectilinear/diagonal bases.
type BB84PhotonDetectHandler struct{}

func (h *BB84PhotonDetectHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"raw_pulses_received":50000000,"sifted_bits_count":1824000,"dark_counts_per_sec":120,"timing_jitter_ps":45}`),
	}, nil
}

// QBERSiftingCheckHandler calculates quantum bit error rate to verify no eavesdropper (Eve) interception.
type QBERSiftingCheckHandler struct{}

func (h *QBERSiftingCheckHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"qber_pct":2.14,"max_tolerable_qber_pct":11.0,"channel_security":"PROVABLY_SECURE","eavesdropper_detected":false}`),
	}, nil
}

// CascadeErrorCorrectionHandler reconciles bit parity errors over authenticated classical channel.
type CascadeErrorCorrectionHandler struct{}

func (h *CascadeErrorCorrectionHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"cascade_passes":4,"bits_corrected":3900,"shannon_efficiency_f":1.12,"residual_bit_errors":0}`),
	}, nil
}

// ToeplitzPrivacyAmpHandler compresses reconciled key using universal hash functions to eliminate Eve's mutual information.
type ToeplitzPrivacyAmpHandler struct{}

func (h *ToeplitzPrivacyAmpHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"final_secure_key_bits":1048576,"hash_function":"TOEPLITZ_MATRIX","key_id":"QKD-KEY-2026-06-02-A","security_parameter":1e-10,"status":"KEYS_COMMITTED_TO_HSM"}`),
	}, nil
}
