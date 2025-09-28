package order

import (
	"context"
	"fmt"
	"route256/loms/internal/domain"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) OrderCancel(ctx context.Context, orderID int64) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderService.OrderCancel")
	defer span.Finish()

	err := s.txManagerMaster.ReadCommitted(ctx, func(ctx context.Context) error {
		order, err := s.orderRepository.GetByOrderIDForUpdate(ctx, orderID)
		if err != nil {
			return fmt.Errorf("orderRepository.GetByOrderIDForUpdate: %w", err)
		}

		switch order.Status {
		case domain.OrderStatusCancelled:
			return nil
		case domain.OrderStatusFailed, domain.OrderStatusPayed:
			return domain.ErrCancelOrder
		}

		if err = s.stockService.ReserveCancel(ctx, order.Items); err != nil {
			return fmt.Errorf("stockService.ReserveRemove: %w", err)
		}

		if err = s.setStatusAndCreateEvent(ctx, orderID, domain.OrderStatusCancelled); err != nil {
			return fmt.Errorf("setStatusAndCreateEvent: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("txManager.ReadCommitted: %w", err)
	}

	return nil
}
