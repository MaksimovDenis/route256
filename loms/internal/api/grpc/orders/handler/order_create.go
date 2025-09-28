package api

import (
	"context"
	"errors"
	"fmt"
	"route256/loms/internal/api/grpc/orders/handler/utils"
	"route256/loms/internal/domain"
	desc "route256/loms/internal/pb/loms/v1"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (hdl *Implementation) OrderCreate(
	ctx context.Context, req *desc.OrderCreateRequest) (
	*desc.OrderCreateResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.OrderCreate")
	defer span.Finish()

	order := mapOrderCreateRequestToDomain(req)

	if err := validateUniqueSkus(order.Items); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	orderID, err := hdl.orderService.OrderCreate(ctx, order)
	if err != nil {
		if errors.Is(err, domain.ErrNotEnoughStock) || errors.Is(err, domain.ErrStockNotFound) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}

		return nil, status.Error(codes.Internal, domain.ErrInternalServerError.Error())
	}

	return &desc.OrderCreateResponse{
		OrderId: orderID,
	}, nil
}

func mapOrderCreateRequestToDomain(req *desc.OrderCreateRequest) domain.Order {
	return domain.Order{
		UserID: req.GetUserId(),
		Items:  utils.MapItemsToDomain(req.GetItems()),
	}
}

func validateUniqueSkus(items []domain.Item) error {
	seen := make(map[domain.Sku]struct{}, len(items))

	for _, item := range items {
		if _, exists := seen[item.Sku]; exists {
			return fmt.Errorf("duplicate SKU in order: %d", item.Sku)
		}
		seen[item.Sku] = struct{}{}
	}
	return nil
}
