package grpc

import (
	"context"
	"time"
)

type UnaryServerInterceptor func(ctx context.Context, req interface{}, handler func(context.Context, interface{}) (interface{}, error)) (interface{}, error)

func LoggingInterceptor(logger func(msg string)) UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, handler func(context.Context, interface{}) (interface{}, error)) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		dur := time.Since(start)
		if logger != nil {
			logger("grpc call completed in " + dur.String())
		}
		return resp, err
	}
}
