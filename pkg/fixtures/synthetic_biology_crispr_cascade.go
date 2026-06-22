package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// SyntheticBiologyCRISPRCascadePipeline builds a multi-guide CRISPR-Cas12a genome engineering validation DAG.
func SyntheticBiologyCRISPRCascadePipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_crispr_cas12a_cascade"),
		TenantID:    "synthetic-genomics-foundry",
		Name:        "Multiplexed CRISPR-Cas12a Transcriptional Activation & Base Editing Validation DAG",
		Version:     1,
		Description: "Simulates crRNA array processing, PAM site TTTV verification, off-target mismatch CFD scoring, and dCas12a-VPR transcriptional upregulation in mammalian cell cultures.",
		Timeout:     90 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "crrna-array-spacer-processing",
				TaskType: "crrna_processing",
			},
			{
				ID:        "pam-motif-verification",
				TaskType:  "pam_verification",
				DependsOn: []string{"crrna-array-spacer-processing"},
			},
			{
				ID:        "off-target-mismatch-cfd-scoring",
				TaskType:  "cfd_offtarget_score",
				DependsOn: []string{"pam-motif-verification"},
			},
			{
				ID:        "transcriptional-activation-assay",
				TaskType:  "vpr_activation_assay",
				DependsOn: []string{"off-target-mismatch-cfd-scoring"},
			},
		},
	}
}

// CrRNAProcessingHandler checks direct repeat maturation by RNase III / Cas12a intrinsic activity.
type CrRNAProcessingHandler struct{}

func (h *CrRNAProcessingHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"num_spacers":6,"direct_repeat_length_nt":19,"processing_efficiency":0.965}`),
	}, nil
}

// PAMVerificationHandler checks presence of 5'-TTTV canonical Cas12a PAM site.
type PAMVerificationHandler struct{}

func (h *PAMVerificationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"pam_valid":true,"pam_sequence":"TTTA","orientation":"5_prime_proximal"}`),
	}, nil
}

// CFDOffTargetScoreHandler computes Cutting Frequency Determination (CFD) off-target matrix scores across genome.
type CFDOffTargetScoreHandler struct{}

func (h *CFDOffTargetScoreHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"aggregate_cfd_score":0.0024,"max_single_offtarget_cfd":0.0008,"cleavage_specificity_passed":true}`),
	}, nil
}

// VPRActivationAssayHandler measures fold induction of target promoters using qRT-PCR / luciferase signal.
type VPRActivationAssayHandler struct{}

func (h *VPRActivationAssayHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"fold_activation":18.4,"target_gene":"OCT4","reporter_luminescence_rlu":45200,"cv_percent":2.1}`),
	}, nil
}
