//go:build integration
// +build integration

package repository_test

import (
	"encoding/json"
	"route256/loms/internal/domain"
	"strconv"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

func (s *Suite) TestCreateEvent_Success(t provider.T) {
	t.Parallel()

	t.Title("Successful order creation")

	testOrderID := int64(1)
	testKey := strconv.FormatInt(testOrderID, 10)

	testPayload := domain.OrderEvent{
		OrderID: testOrderID,
		Status:  string(domain.OrderStatusNew),
		Moment:  time.Now(),
	}

	payload, err := json.Marshal(testPayload)
	t.Require().NoError(err)

	testEvent := domain.Event{
		ID:      0,
		Topic:   "test_topic",
		Key:     testKey,
		Payload: payload,
	}

	t.WithNewStep("create order", func(sCtx provider.StepCtx) {
		err := s.outboxRepo.CreateEvent(s.ctx, testEvent)
		sCtx.Require().NoError(err)
	})

	t.WithNewStep("fetch next message", func(sCtx provider.StepCtx) {
		events, err := s.outboxRepo.FetchNextMessages(s.ctx, 100)
		sCtx.Require().NoError(err)

		var actualEvent domain.Event

		for idx, event := range events {
			if event.Key == testKey {
				actualEvent = events[idx]
			}
		}

		sCtx.Require().Equal(testEvent.Topic, actualEvent.Topic)
		sCtx.Require().Equal(testEvent.Key, actualEvent.Key)
		sCtx.Require().Equal(domain.EventStatusNew, actualEvent.Status)
	})
}
