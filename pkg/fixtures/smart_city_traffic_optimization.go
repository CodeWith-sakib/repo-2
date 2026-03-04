package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// SmartCityTrafficOptimizationPipeline builds a metropolitan traffic signal coordination & congestion control DAG.
func SmartCityTrafficOptimizationPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_smart_city_traffic"),
		TenantID:    "metropolitan-transit-authority",
		Name:        "Metropolitan Real-Time Traffic Signal Optimization & Incident Response Pipeline",
		Version:     1,
		Description: "Ingests induction loop sensors and CCTV video feeds, computes intersection vehicle queue lengths, optimizes split/cycle times via Webster's formula, calculates green wave transit priority corridors, and broadcasts dynamic VMS signage updates.",
		Timeout:     60 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "loop-and-cctv-telemetry-ingest",
				TaskType: "traffic_telemetry_ingest",
			},
			{
				ID:        "queue-length-estimation",
				TaskType:  "queue_length_estimator",
				DependsOn: []string{"loop-and-cctv-telemetry-ingest"},
			},
			{
				ID:        "webster-signal-cycle-optimizer",
				TaskType:  "webster_cycle_optimizer",
				DependsOn: []string{"queue-length-estimation"},
			},
			{
				ID:        "green-wave-transit-priority",
				TaskType:  "green_wave_corridor",
				DependsOn: []string{"webster-signal-cycle-optimizer"},
			},
			{
				ID:        "dynamic-vms-signage-broadcast",
				TaskType:  "vms_signage_broadcast",
				DependsOn: []string{"green-wave-transit-priority"},
			},
		},
	}
}

// TrafficTelemetryHandler reads magnetic induction loops and vehicle counts.
type TrafficTelemetryHandler struct{}

func (h *TrafficTelemetryHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"corridor_id":"BROADWAY-CORRIDOR-7","intersections_reporting":24,"average_speed_kmh":28.5,"occupancy_pct":78.2}`),
	}, nil
}

// QueueLengthEstimatorHandler estimates vehicle spillover queuing per intersection arm.
type QueueLengthEstimatorHandler struct{}

func (h *QueueLengthEstimatorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"max_queue_vehicles":42,"critical_intersection":"BROADWAY-42ND","spillover_risk":"MODERATE"}`),
	}, nil
}

// WebsterCycleOptimizerHandler derives optimal cycle length and phase green splits.
type WebsterCycleOptimizerHandler struct{}

func (h *WebsterCycleOptimizerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"optimal_cycle_seconds":110,"phase_a_green_sec":55,"phase_b_green_sec":45,"lost_time_sec":10}`),
	}, nil
}

// GreenWaveCorridorHandler calculates phase offsets to guarantee uninterrupted transit vehicle green waves.
type GreenWaveCorridorHandler struct{}

func (h *GreenWaveCorridorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"progression_speed_kmh":35.0,"bandwidth_seconds":28,"transit_buses_prioritized":6}`),
	}, nil
}

// VMSSignageBroadcastHandler pushes travel time advisories to variable message signs.
type VMSSignageBroadcastHandler struct{}

func (h *VMSSignageBroadcastHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"vms_displays_updated":12,"display_message":"EST DOWNTOWN TRAVEL TIME: 14 MIN - SPEED ADVISORY: 35 KM/H","status":"BROADCASTED"}`),
	}, nil
}
