package order

import (
	"context"
	"encoding/json"
	"fmt"
	"route256/loms/internal/domain"
	"strconv"
	"time"

	"github.com/opentracing/opentracing-go"
)

type orderEvent struct {
	OrderID int64     `json:"order_id"`
	Status  string    `json:"status"`
	Moment  time.Time `json:"moment"`
}

func (s *Service) prepareOrderEvent(orderID int64, status domain.OrderStatus) (domain.Event, error) {
	payload := orderEvent{
		OrderID: orderID,
		Status:  string(status),
		Moment:  time.Now(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return domain.Event{}, fmt.Errorf("json.Marshal: failed to marshal payload: %w", err)
	}

	event := domain.Event{
		Topic:   s.orderTopic,
		Key:     strconv.FormatInt(orderID, 10),
		Payload: data,
		Status:  domain.EventStatusNew,
	}

	return event, nil
}

func (s *Service) createEvent(ctx context.Context, orderID int64, status domain.OrderStatus) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderService.createEvent")
	defer span.Finish()

	event, err := s.prepareOrderEvent(orderID, status)
	if err != nil {
		return fmt.Errorf("prepareOrderEvent: %w", err)
	}

	if repoErr := s.eventRepository.CreateEvent(ctx, event); repoErr != nil {
		return fmt.Errorf("eventRepository.CreateEventt: %w", err)
	}

	return nil
}

func (s *Service) setStatusAndCreateEvent(ctx context.Context, orderID int64, status domain.OrderStatus) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderService.setStatusAndCreateEvent")
	defer span.Finish()

	event, err := s.prepareOrderEvent(orderID, status)
	if err != nil {
		return fmt.Errorf("prepareOrderEvent: %w", err)
	}

	if repoErr := s.orderRepository.SetStatusAndCreateEvent(ctx, orderID, status, event); repoErr != nil {
		return fmt.Errorf("orderRepository.SetStatusAndCreateEvent: %w", repoErr)
	}

	return nil
}
