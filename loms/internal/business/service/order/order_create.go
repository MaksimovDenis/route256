package order

import (
	"context"
	"errors"
	"fmt"
	"route256/loms/internal/domain"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) OrderCreate(ctx context.Context, order domain.Order) (int64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderService.OrderCreate")
	defer span.Finish()

	var orderID int64

	if err := s.txManagerMaster.ReadCommitted(ctx, func(ctx context.Context) error {
		var err error
		orderID, err = s.orderRepository.CreateOrder(ctx, order.UserID)
		if err != nil {
			return fmt.Errorf("orderRepository.CreateOrder: %w", err)
		}

		if err := s.orderRepository.CreateOrderItems(ctx, orderID, order.Items); err != nil {
			return fmt.Errorf("orderRepository.CreateOrderItems: %w", err)
		}

		if err := s.createEvent(ctx, orderID, domain.OrderStatusNew); err != nil {
			return fmt.Errorf("createEvent: %w", err)
		}

		return nil
	}); err != nil {
		return 0, fmt.Errorf("OrderCreate failed: %w", err)
	}

	if err := s.txManagerMaster.ReadCommitted(ctx, func(ctx context.Context) error {
		if err := s.stockService.Reserve(ctx, order.Items); err != nil {
			return fmt.Errorf("stockService.Reserve: %w", err)
		}

		if err := s.setStatusAndCreateEvent(ctx, orderID, domain.OrderStatusAwaitingPayment); err != nil {
			return fmt.Errorf("setStatusAndCreateEvent: %w", err)
		}

		return nil
	}); err != nil {
		if errors.Is(err, domain.ErrNotEnoughStock) || errors.Is(err, domain.ErrStockNotFound) {
			if errStatus := s.setStatusAndCreateEvent(ctx, orderID, domain.OrderStatusFailed); errStatus != nil {
				return 0, fmt.Errorf("setStatusAndCreateEvent: %w", errStatus)
			}
		}

		return 0, fmt.Errorf("OrderCreate: stock reservation failed: %w", err)
	}

	return orderID, nil
}
