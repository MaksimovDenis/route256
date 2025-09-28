package order

import (
	"context"
	"fmt"
	"route256/loms/internal/domain"
	"sort"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) OrderInfo(ctx context.Context, orderID int64) (domain.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderService.OrderInfo")
	defer span.Finish()

	order, err := s.orderRepository.GetByOrderID(ctx, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("orderRepository.GetByOrderID: %w", err)
	}

	sort.Slice(order.Items, func(i, j int) bool {
		return order.Items[i].Sku < order.Items[j].Sku
	})

	return order, nil
}
