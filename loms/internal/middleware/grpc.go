package middleware

import (
	"context"
	"route256/loms/internal/infra/metrics"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const traceIDKey = "x-trace-id"

func Validate(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if v, ok := req.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	return handler(ctx, req)
}

func MetricsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	metrics.IncRequestCounter()

	start := time.Now()

	res, err := handler(ctx, req)
	duration := time.Since(start).Seconds()

	if err != nil {
		metrics.IncGrpcRequestCounter(info.FullMethod, "error")
		metrics.GrpcRequestDurationHistogram(info.FullMethod, "error", duration)
	} else {
		metrics.IncGrpcRequestCounter(info.FullMethod, "success")
		metrics.GrpcRequestDurationHistogram(info.FullMethod, "success", duration)
	}

	return res, err
}

func ServerTracingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, info.FullMethod)
	defer span.Finish()

	spanContext, ok := span.Context().(jaeger.SpanContext)
	if ok {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(traceIDKey, spanContext.TraceID().String()))

		header := metadata.New(map[string]string{traceIDKey: spanContext.TraceID().String()})
		err := grpc.SendHeader(ctx, header)
		if err != nil {
			return nil, err
		}
	}

	res, err := handler(ctx, req)
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("err", err.Error())
	}

	return res, err
}
