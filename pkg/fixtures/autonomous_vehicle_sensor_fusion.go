package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// AutonomousVehicleSensorFusionPipeline builds an automotive perception & sensor fusion DAG.
func AutonomousVehicleSensorFusionPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_av_sensor_fusion"),
		TenantID:    "automotive-perception-ai",
		Name:        "Automotive Multi-Camera LiDAR Radar Sensor Fusion & Motion Planning Pipeline",
		Version:     1,
		Description: "Synchronizes 32-beam LiDAR point clouds with 8 HDR surround cameras, detects 3D bounding boxes via PointPillars, associates tracks using Extended Kalman Filter (EKF), predicts 5-second road agent trajectory heatmaps, and generates collision-free spline vehicle control trajectories.",
		Timeout:     45 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "lidar-camera-timesync-ingest",
				TaskType: "sensor_timesync_ingest",
			},
			{
				ID:        "pointpillars-3d-detector",
				TaskType:  "pointpillars_3d_detect",
				DependsOn: []string{"lidar-camera-timesync-ingest"},
			},
			{
				ID:        "ekf-multitarget-track-association",
				TaskType:  "ekf_track_association",
				DependsOn: []string{"pointpillars-3d-detector"},
			},
			{
				ID:        "trajectory-heatmap-prediction",
				TaskType:  "trajectory_heatmap_pred",
				DependsOn: []string{"ekf-multitarget-track-association"},
			},
			{
				ID:        "quintic-spline-motion-planner",
				TaskType:  "quintic_spline_planner",
				DependsOn: []string{"trajectory-heatmap-prediction"},
			},
		},
	}
}

// SensorTimeSyncHandler synchronizes sensor frames across PTP IEEE 1588 hardware clocks.
type SensorTimeSyncHandler struct{}

func (h *SensorTimeSyncHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"lidar_points_count":128000,"cameras_synchronized":8,"timesync_jitter_us":12,"vehicle_speed_mps":22.4}`),
	}, nil
}

// PointPillarsDetectorHandler extracts 3D voxel features and generates oriented bounding boxes.
type PointPillarsDetectorHandler struct{}

func (h *PointPillarsDetectorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"vehicles_detected":14,"pedestrians_detected":3,"cyclists_detected":1,"inference_time_ms":14.2}`),
	}, nil
}

// EKFTrackAssociationHandler performs Mahalanobis distance gating and Hungarian data association.
type EKFTrackAssociationHandler struct{}

func (h *EKFTrackAssociationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"active_tracks":18,"covariance_norm":0.042,"lost_tracks":0,"closest_obstacle_distance_m":18.5}`),
	}, nil
}

// TrajectoryHeatmapPredictionHandler predicts multi-modal agent intentions using graph neural networks.
type TrajectoryHeatmapPredictionHandler struct{}

func (h *TrajectoryHeatmapPredictionHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"prediction_horizon_sec":5.0,"lane_change_probabilities":[0.02,0.95,0.03],"collision_risk_score":0.01}`),
	}, nil
}

// QuinticSplinePlannerHandler generates jerk-minimized C2-continuous motion profiles.
type QuinticSplinePlannerHandler struct{}

func (h *QuinticSplinePlannerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"steering_angle_deg":-2.4,"throttle_pct":18.0,"brake_bar":0.0,"trajectory_status":"OPTIMAL_FEASIBLE"}`),
	}, nil
}
