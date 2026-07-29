package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// OrbitalDebrisCollisionConjunctionPipeline builds a space situational awareness (SSA) conjunction assessment DAG.
func OrbitalDebrisCollisionConjunctionPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_orbital_conjunction_assessment"),
		TenantID:    "space-traffic-management-agency",
		Name:        "LEO Satellite Mega-Constellation Orbital Debris Conjunction & Collision Avoidance DAG",
		Version:     1,
		Description: "Ingests Two-Line Element (TLE) ephemerides, propagates orbits using SGP4 perturbed numerical integration, constructs 3D covariance collision ellipsoids via Foster 1992 method, computes Probability of Collision (Pc), and plans optimal delta-V thrust maneuvers.",
		Timeout:     90 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "tle-orbit-propagation-sgp4",
				TaskType: "sgp4_orbit_propagation",
			},
			{
				ID:        "close-approach-distance-screening",
				TaskType:  "conjunction_distance_screen",
				DependsOn: []string{"tle-orbit-propagation-sgp4"},
			},
			{
				ID:        "covariance-collision-probability-calc",
				TaskType:  "collision_probability_calc",
				DependsOn: []string{"close-approach-distance-screening"},
			},
			{
				ID:        "optimal-avoidance-deltav-planner",
				TaskType:  "avoidance_deltav_planner",
				DependsOn: []string{"covariance-collision-probability-calc"},
			},
		},
	}
}

// SGP4OrbitPropagationHandler propagates state vectors under J2-J4 geopotential and atmospheric drag.
type SGP4OrbitPropagationHandler struct{}

func (h *SGP4OrbitPropagationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"primary_sat_id":"SAT-44921","debris_norad_id":"DEB-1982-092A","orbital_altitude_km":540.2,"relative_velocity_km_s":14.35}`),
	}, nil
}

// ConjunctionDistanceScreenHandler performs bounding-box temporal sieve to find Time of Closest Approach (TCA).
type ConjunctionDistanceScreenHandler struct{}

func (h *ConjunctionDistanceScreenHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"time_to_closest_approach_sec":4320.0,"miss_distance_radial_m":12.5,"miss_distance_in_track_m":48.0,"miss_distance_cross_track_m":8.2,"total_miss_distance_m":50.2}`),
	}, nil
}

// CollisionProbabilityCalcHandler integrates 2D projected collision ellipse using Chan/Foster method.
type CollisionProbabilityCalcHandler struct{}

func (h *CollisionProbabilityCalcHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"combined_hard_body_radius_m":5.0,"probability_of_collision_pc":0.00342,"conjunction_warning_level":"RED_ALERT"}`),
	}, nil
}

// AvoidanceDeltaVPlannerHandler designs minimum fuel burns perpendicular to collision plane.
type AvoidanceDeltaVPlannerHandler struct{}

func (h *AvoidanceDeltaVPlannerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"burn_delta_v_m_s":0.185,"burn_direction":"ANTI_VELOCITY","projected_post_burn_miss_distance_m":2450.0,"maneuver_decision":"EXECUTE_AVOIDANCE"}`),
	}, nil
}
