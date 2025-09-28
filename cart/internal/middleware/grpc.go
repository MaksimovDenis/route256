package middleware

import (
	"context"
	"route256/cart/internal/infra/metrics"
	"time"

	"google.golang.org/grpc"
)

func MetricsClientInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	start := time.Now()

	err := invoker(ctx, method, req, reply, cc, opts...)
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}

	metrics.IncExternalRequestCounter(method, status)
	metrics.ExternalRequestDurationHistogram(method, status, duration)

	return err
}
