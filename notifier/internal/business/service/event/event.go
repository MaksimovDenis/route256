package event

import (
	"context"
	"route256/notifier/internal/domain"
	"route256/notifier/internal/infra/logger"
)

func (s *Service) ProcessOrderEvent(ctx context.Context, event domain.OrderEvent) error {
	date := event.Moment.Format("2006-01-02 15:04:05")
	logger.Infof(ctx, "order_id = %v, status = %s, moment = %v", event.OrderID, event.Status, date)

	return nil
}
