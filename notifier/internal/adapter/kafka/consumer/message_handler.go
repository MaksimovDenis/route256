package consumer

import (
	"context"
	"encoding/json"
	"route256/notifier/internal/domain"
	"route256/notifier/internal/infra/logger"
	"time"

	"github.com/IBM/sarama"
)

type eventService interface {
	ProcessOrderEvent(ctx context.Context, event domain.OrderEvent) error
}
type GroupHandler struct {
	eventService eventService
}

type msgOrderEvent struct {
	OrderID int64     `json:"order_id"`
	Status  string    `json:"status"`
	Moment  time.Time `json:"moment"`
}

func NewGroupHandler(eventService eventService) *GroupHandler {
	return &GroupHandler{eventService: eventService}
}

func (c *GroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *GroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *GroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := context.Background()

	for message := range claim.Messages() {
		event := msgOrderEvent{}
		if err := json.Unmarshal(message.Value, &event); err != nil {
			logger.Errorf(ctx, "failed to unmarshal message: %v", err)
			session.MarkMessage(message, "")
			continue
		}

		if err := c.eventService.ProcessOrderEvent(session.Context(), mapOrderEvent(event)); err != nil {
			logger.Errorf(ctx, "eventService.ProcessOrderEvent: error handling domain event: %v", err)
			continue
		}

		session.MarkMessage(message, "")
	}

	logger.Infof(ctx, "message channel was closed")
	return nil
}

func mapOrderEvent(in msgOrderEvent) domain.OrderEvent {
	return domain.OrderEvent{
		OrderID: in.OrderID,
		Status:  in.Status,
		Moment:  in.Moment,
	}
}
