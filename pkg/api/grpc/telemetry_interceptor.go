package grpc

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// MethodTelemetry holds runtime latency and invocation metrics for a gRPC method.
type MethodTelemetry struct {
	TotalCalls atomic.Int64
	Successes  atomic.Int64
	Failures   atomic.Int64
	Panics     atomic.Int64
}

// GRPCServerTelemetry records gRPC invocation metrics.
type GRPCServerTelemetry struct {
	mu      sync.RWMutex
	methods map[string]*MethodTelemetry
}

// NewGRPCServerTelemetry initializes gRPC telemetry recorder.
func NewGRPCServerTelemetry() *GRPCServerTelemetry {
	return &GRPCServerTelemetry{
		methods: make(map[string]*MethodTelemetry),
	}
}

func (t *GRPCServerTelemetry) getMethod(method string) *MethodTelemetry {
	t.mu.RLock()
	m, ok := t.methods[method]
	t.mu.RUnlock()

	if !ok {
		t.mu.Lock()
		if m2, ok2 := t.methods[method]; ok2 {
			m = m2
		} else {
			m = &MethodTelemetry{}
			t.methods[method] = m
		}
		t.mu.Unlock()
	}
	return m
}

// TelemetryInterceptor returns a UnaryServerInterceptor that measures duration, updates stats, and recovers panics.
func (t *GRPCServerTelemetry) TelemetryInterceptor(methodName string) UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, handler func(context.Context, interface{}) (interface{}, error)) (resp interface{}, err error) {
		m := t.getMethod(methodName)
		m.TotalCalls.Add(1)

		defer func() {
			if r := recover(); r != nil {
				m.Panics.Add(1)
				m.Failures.Add(1)
				err = fmt.Errorf("panic recovered in gRPC handler: %v", r)
			}
		}()

		resp, err = handler(ctx, req)
		if err != nil {
			m.Failures.Add(1)
		} else {
			m.Successes.Add(1)
		}

		return resp, err
	}
}

// GetStats returns call metrics for a method name.
func (t *GRPCServerTelemetry) GetStats(method string) (calls, succ, fail, panics int64) {
	m := t.getMethod(method)
	return m.TotalCalls.Load(), m.Successes.Load(), m.Failures.Load(), m.Panics.Load()
}
