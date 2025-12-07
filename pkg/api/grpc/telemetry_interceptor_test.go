package grpc

import (
	"context"
	"fmt"
	"testing"
)

func TestGRPCServerTelemetry_SuccessAndFailure(t *testing.T) {
	telemetry := NewGRPCServerTelemetry()
	method := "/kestrel.v1.WorkflowService/CreateRun"
	interceptor := telemetry.TelemetryInterceptor(method)

	// 1. Successful call
	handlerSuccess := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}
	_, err := interceptor(context.Background(), "req", handlerSuccess)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2. Errored call
	handlerError := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, fmt.Errorf("run not found")
	}
	_, err = interceptor(context.Background(), "req", handlerError)
	if err == nil {
		t.Fatal("expected error from error handler")
	}

	calls, succ, fail, _ := telemetry.GetStats(method)
	if calls != 2 || succ != 1 || fail != 1 {
		t.Errorf("stats mismatch: calls=%d succ=%d fail=%d", calls, succ, fail)
	}
}

func TestGRPCServerTelemetry_PanicRecovery(t *testing.T) {
	telemetry := NewGRPCServerTelemetry()
	method := "/kestrel.v1.WorkflowService/PanicMethod"
	interceptor := telemetry.TelemetryInterceptor(method)

	handlerPanic := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("fatal nil pointer dereference")
	}

	_, err := interceptor(context.Background(), "req", handlerPanic)
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}

	_, _, _, panics := telemetry.GetStats(method)
	if panics != 1 {
		t.Errorf("expected 1 panic recorded, got %d", panics)
	}
}
