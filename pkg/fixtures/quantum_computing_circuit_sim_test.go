package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestQuantumCircuitSimulationPipeline_Validate(t *testing.T) {
	wf := QuantumCircuitSimulationPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestQuantumCircuitSimulationPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-quantum-01", StepID: "openqasm-circuit-parse"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"qasm_parse", &QASMParseHandler{}},
		{"pauli_synth", &PauliSynthHandler{}},
		{"statevector_sim", &StatevectorSimHandler{}},
		{"zne_mitigation", &ZNEMitigationHandler{}},
		{"vqe_optimizer", &VQEOptimizerHandler{}},
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
