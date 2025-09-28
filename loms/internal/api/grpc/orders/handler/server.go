package api

import (
	"context"
	"route256/loms/internal/domain"
	desc "route256/loms/internal/pb/loms/v1"
)

type orderService interface {
	OrderCreate(ctx context.Context, order domain.Order) (int64, error)
	OrderInfo(ctx context.Context, orderID int64) (domain.Order, error)
	OrderPay(ctx context.Context, orderID int64) error
	OrderCancel(ctx context.Context, orderID int64) error
}

type stockService interface {
	StocksInfo(ctx context.Context, sku domain.Sku) (int64, error)
}

type Implementation struct {
	desc.UnimplementedOrdersServer
	desc.UnimplementedStocksServer
	desc.UnimplementedHealthServer
	orderService orderService
	stockService stockService
}

func NewImplementation(orderService orderService, stockService stockService) *Implementation {
	return &Implementation{
		orderService: orderService,
		stockService: stockService,
	}
}
