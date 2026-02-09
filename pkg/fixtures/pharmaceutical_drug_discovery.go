package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// PharmaceuticalDrugDiscoveryPipeline builds an automated structure-based virtual screening DAG.
func PharmaceuticalDrugDiscoveryPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_drug_discovery_screen"),
		TenantID:    "oncology-therapeutics-lab",
		Name:        "Structure-Based Virtual Screening & Molecular Docking Pipeline",
		Version:     1,
		Description: "Fetches target kinase PDB crystal structure, performs protonation/energy minimization, docks billion-compound ZINC20 SMILES library via AutoDock Vina, computes MM-GBSA binding free energy, and runs ADMET toxicity screening.",
		Timeout:     240 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "target-kinase-pdb-prep",
				TaskType: "target_pdb_prep",
			},
			{
				ID:        "conformational-energy-minimization",
				TaskType:  "energy_minimization",
				DependsOn: []string{"target-kinase-pdb-prep"},
			},
			{
				ID:        "autodock-vina-virtual-docking",
				TaskType:  "autodock_docking",
				DependsOn: []string{"conformational-energy-minimization"},
			},
			{
				ID:        "mm-gbsa-free-energy-scoring",
				TaskType:  "mm_gbsa_scoring",
				DependsOn: []string{"autodock-vina-virtual-docking"},
			},
			{
				ID:        "admet-pharmacokinetics-screening",
				TaskType:  "admet_screening",
				DependsOn: []string{"mm-gbsa-free-energy-scoring"},
			},
		},
	}
}

// TargetPDBPrepHandler cleans solvent, adds polar hydrogens, and computes Gasteiger partial charges.
type TargetPDBPrepHandler struct{}

func (h *TargetPDBPrepHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"pdb_id":"6VXX","resolution_angstrom":2.4,"active_site_residues":["ASP-166","LYS-89","PHE-185"],"status":"PREPARED"}`),
	}, nil
}

// EnergyMinimizationHandler applies CHARMM36 force field minimization.
type EnergyMinimizationHandler struct{}

func (h *EnergyMinimizationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"force_field":"CHARMM36m","rmsd_angstrom":0.18,"potential_energy_kcal_mol":-14250.2}`),
	}, nil
}

// AutoDockDockingHandler computes rigid receptor, flexible ligand Monte Carlo docking poses.
type AutoDockDockingHandler struct{}

func (h *AutoDockDockingHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"compounds_screened":100000,"top_hits_count":50,"best_affinity_kcal_mol":-11.8,"exhaustiveness":16}`),
	}, nil
}

// MMGBSAScoringHandler calculates continuum solvent molecular mechanics binding free energies.
type MMGBSAScoringHandler struct{}

func (h *MMGBSAScoringHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"delta_g_bind_kcal_mol":-58.4,"van_der_waals_kcal_mol":-64.2,"electrostatic_kcal_mol":-18.9,"entropy_t_delta_s":24.7}`),
	}, nil
}

// ADMETScreeningHandler screens for hERG cardiac channel inhibition, blood-brain barrier penetration, and CYP3A4 clearance.
type ADMETScreeningHandler struct{}

func (h *ADMETScreeningHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"lead_candidate_id":"LIG-CHEMBL-44910","herg_cardiotoxicity":"NEGATIVE","cyp3a4_inhibition":false,"lipinski_rule_of_5_violations":0,"status":"CANDIDATE_APPROVED"}`),
	}, nil
}
