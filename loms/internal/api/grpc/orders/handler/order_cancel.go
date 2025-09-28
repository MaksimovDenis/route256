package api

import (
	"context"
	"errors"
	"route256/loms/internal/domain"
	desc "route256/loms/internal/pb/loms/v1"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (hdl *Implementation) OrderCancel(ctx context.Context, req *desc.OrderCancelRequest) (*desc.OrderCancelResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.OrderCancel")
	defer span.Finish()

	err := hdl.orderService.OrderCancel(ctx, req.GetOrderId())
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		} else if errors.Is(err, domain.ErrCancelOrder) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}

		return nil, status.Error(codes.Internal, domain.ErrInternalServerError.Error())
	}

	return &desc.OrderCancelResponse{}, nil
}
