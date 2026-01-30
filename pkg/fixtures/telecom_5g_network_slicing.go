package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// Telecom5GNetworkSlicingPipeline builds a 3GPP 5G Core Network Slicing & RAN Orchestration DAG.
func Telecom5GNetworkSlicingPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_5g_slice_orchestration"),
		TenantID:    "telecom-core-ops",
		Name:        "3GPP 5G Standalone (SA) Network Slice Lifecycle & SLA Assurance DAG",
		Version:     1,
		Description: "Ingests gNodeB CU/DU telemetry, admits Ultra-Reliable Low-Latency Communication (URLLC) slice, dynamically configures User Plane Function (UPF) DPDK data paths, programs SDN OpenFlow forwarding tables, and monitors E2E latency SLA guarantees.",
		Timeout:     90 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "ran-cudu-telemetry-ingest",
				TaskType: "cudu_telemetry_ingest",
			},
			{
				ID:        "slice-admission-urllc-check",
				TaskType:  "slice_admission_control",
				DependsOn: []string{"ran-cudu-telemetry-ingest"},
			},
			{
				ID:        "upf-dpdk-dataplane-configure",
				TaskType:  "upf_dpdk_configure",
				DependsOn: []string{"slice-admission-urllc-check"},
			},
			{
				ID:        "sdn-openflow-routing-program",
				TaskType:  "sdn_routing_program",
				DependsOn: []string{"upf-dpdk-dataplane-configure"},
			},
			{
				ID:        "e2e-sla-latency-assurance",
				TaskType:  "sla_latency_assurance",
				DependsOn: []string{"sdn-openflow-routing-program"},
			},
		},
	}
}

// CUDUTelemetryHandler parses O-RAN Fronthaul and Midhaul performance indicators.
type CUDUTelemetryHandler struct{}

func (h *CUDUTelemetryHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"gnb_id":"GNB-METRO-042","active_prb_utilization_pct":64.8,"cudu_rtt_us":210,"packet_loss_rate":0.00002}`),
	}, nil
}

// SliceAdmissionHandler verifies bandwidth and QoS class identifier (5QI) feasibility.
type SliceAdmissionHandler struct{}

func (h *SliceAdmissionHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"slice_nssai":{"sst":1,"sd":"000102"},"5qi":82,"guaranteed_flow_bitrate_mbps":150,"admission":"APPROVED"}`),
	}, nil
}

// UPFDPDKHandler provisions hugepages and fast-path kernel bypass interfaces.
type UPFDPDKHandler struct{}

func (h *UPFDPDKHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"upf_instance":"UPF-EDGE-01","dpdk_rx_tx_cores":4,"flow_table_entries":100000,"status":"PROGRAMMED"}`),
	}, nil
}

// SDNRoutingHandler pushes OpenFlow flow-mod entries to spine-leaf data fabric.
type SDNRoutingHandler struct{}

func (h *SDNRoutingHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"sdn_controller":"ONOS-CLUSTER","flows_installed":128,"vxlan_tunnel_id":4096,"status":"FLOWS_ACTIVE"}`),
	}, nil
}

// SLALatencyAssuranceHandler verifies packet delay variation (PDV) remains below 1 millisecond.
type SLALatencyAssuranceHandler struct{}

func (h *SLALatencyAssuranceHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"measured_p99_latency_ms":0.74,"target_sla_ms":1.0,"jitter_us":45,"sla_conformance":"PASS"}`),
	}, nil
}
