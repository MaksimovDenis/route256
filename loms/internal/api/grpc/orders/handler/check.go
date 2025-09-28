package api

import (
	"context"
	desc "route256/loms/internal/pb/loms/v1"

	"github.com/opentracing/opentracing-go"
)

func (hdl *Implementation) Check(ctx context.Context, _ *desc.HealthCheckRequest) (*desc.HealthCheckResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "api.Check")
	defer span.Finish()

	return &desc.HealthCheckResponse{
		Message: "OK",
	}, nil

}
