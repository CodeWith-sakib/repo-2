package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// QuantumCircuitSimulationPipeline builds a noisy intermediate-scale quantum (NISQ) circuit execution DAG.
func QuantumCircuitSimulationPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_quantum_vqe_sim"),
		TenantID:    "quantum-chem-lab",
		Name:        "Variational Quantum Eigensolver (VQE) Molecular Ground State Simulation DAG",
		Version:     1,
		Description: "Parses OpenQASM 3.0 circuit description, synthesizes Pauli string Hamiltonians for LiH molecule, executes statevector simulation with depolarizing noise model, applies zero-noise extrapolation (ZNE) error mitigation, and optimizes ansatz variational parameters.",
		Timeout:     120 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "openqasm-circuit-parse",
				TaskType: "qasm_parse",
			},
			{
				ID:        "pauli-hamiltonian-synth",
				TaskType:  "pauli_synth",
				DependsOn: []string{"openqasm-circuit-parse"},
			},
			{
				ID:        "noisy-statevector-sim",
				TaskType:  "statevector_sim",
				DependsOn: []string{"pauli-hamiltonian-synth"},
			},
			{
				ID:        "zne-error-mitigation",
				TaskType:  "zne_mitigation",
				DependsOn: []string{"noisy-statevector-sim"},
			},
			{
				ID:        "vqe-parameter-optimizer",
				TaskType:  "vqe_optimizer",
				DependsOn: []string{"zne-error-mitigation"},
			},
		},
	}
}

// QASMParseHandler parses OpenQASM AST into 2-qubit CNOT and 1-qubit U3 gates.
type QASMParseHandler struct{}

func (h *QASMParseHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"circuit_depth":28,"num_qubits":4,"cnot_count":12,"single_qubit_gates":36,"qasm_version":"3.0"}`),
	}, nil
}

// PauliSynthHandler constructs Jordan-Wigner second-quantized fermionic Hamiltonian operators.
type PauliSynthHandler struct{}

func (h *PauliSynthHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"molecule":"LiH","nuclear_repulsion_hartree":0.992,"pauli_terms_count":27,"basis_set":"STO-3G"}`),
	}, nil
}

// StatevectorSimHandler calculates complex state vector amplitudes under Kraus error channels.
type StatevectorSimHandler struct{}

func (h *StatevectorSimHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"shots":10000,"t1_us":85.2,"t2_us":62.4,"depolarizing_rate":0.001,"unmitigated_energy_hartree":-7.8214}`),
	}, nil
}

// ZNEMitigationHandler extrapolates pulse-stretched noise scaling factors to zero error.
type ZNEMitigationHandler struct{}

func (h *ZNEMitigationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"scale_factors":[1.0,1.5,2.0],"extrapolation_model":"RICHARDSON","mitigated_energy_hartree":-7.8821,"error_reduction_pct":78.4}`),
	}, nil
}

// VQEOptimizerHandler runs Broyden-Fletcher-Goldfarb-Shanno (BFGS) gradient optimization.
type VQEOptimizerHandler struct{}

func (h *VQEOptimizerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"iteration":14,"gradient_norm":0.00042,"ground_state_energy_hartree":-7.8823,"chemical_accuracy_reached":true}`),
	}, nil
}
