package api

import (
	"context"
	"errors"
	"route256/loms/internal/api/grpc/orders/handler/utils"
	"route256/loms/internal/domain"
	desc "route256/loms/internal/pb/loms/v1"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (hdl *Implementation) OrderInfo(
	ctx context.Context, req *desc.OrderInfoRequest) (
	*desc.OrderInfoResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.OrderInfo")
	defer span.Finish()

	order, err := hdl.orderService.OrderInfo(ctx, req.GetOrderId())
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, domain.ErrInternalServerError.Error())
	}

	mapItems, err := utils.ItemsDomainToMap(order.Items)
	if err != nil {
		return nil, status.Error(codes.Internal, domain.ErrInternalServerError.Error())
	}

	return &desc.OrderInfoResponse{
		Status: string(order.Status),
		UserId: order.UserID,
		Items:  mapItems,
	}, nil
}
