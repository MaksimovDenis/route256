package orderevent

import (
	"context"
	"fmt"
	"route256/loms/internal/infra/logger"

	"github.com/opentracing/opentracing-go"
)

func (c *CronProcessor) Do(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orederEvent.Do")
	defer span.Finish()

	if err := c.txManagerMaster.ReadCommitted(ctx, func(ctx context.Context) error {
		events, err := c.eventRepository.FetchNextMessages(ctx, c.limitMsg)
		if err != nil {
			return fmt.Errorf("eventRepository.FetchNextMessage: %w", err)
		}

		if len(events) == 0 {
			return nil
		}

		successIDs, errorIDs, err := c.producerKafka.SendOrderEventsBatch(ctx, events)
		if err != nil {
			return err
		}

		if len(successIDs) > 0 {
			if err := c.eventRepository.MarkAsSent(ctx, successIDs); err != nil {
				return fmt.Errorf("eventRepository.MarkAsSent: %w", err)
			}
		}

		if len(errorIDs) > 0 {
			if err := c.eventRepository.MarkAsError(ctx, errorIDs); err != nil {
				return fmt.Errorf("eventRepository.MarkAsError: %w", err)
			}
		}

		return nil
	}); err != nil {
		logger.Errorf(ctx, "eventService.handlePendingEvents: %v", err)
	}

	return nil
}
