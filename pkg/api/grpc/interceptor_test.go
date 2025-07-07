package grpc

import (
	"context"
	"testing"
)

func TestLoggingInterceptor(t *testing.T) {
	logged := false
	interceptor := LoggingInterceptor(func(msg string) {
		logged = true
	})
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}
	_, err := interceptor(context.Background(), "req", handler)
	if err != nil || !logged {
		t.Errorf("interceptor failed: %v, logged: %v", err, logged)
	}
}
