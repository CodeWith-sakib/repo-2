package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestAutonomousVehicleSensorFusionPipeline_Validate(t *testing.T) {
	wf := AutonomousVehicleSensorFusionPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestAutonomousVehicleSensorFusionPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-av-fusion-01", StepID: "lidar-camera-timesync-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"sensor_timesync_ingest", &SensorTimeSyncHandler{}},
		{"pointpillars_3d_detect", &PointPillarsDetectorHandler{}},
		{"ekf_track_association", &EKFTrackAssociationHandler{}},
		{"trajectory_heatmap_pred", &TrajectoryHeatmapPredictionHandler{}},
		{"quintic_spline_planner", &QuinticSplinePlannerHandler{}},
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
