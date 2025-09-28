package cart

import (
	"context"
	"route256/cart/internal/infra/metrics"
)

func (c *CronProcessor) Do(_ context.Context) error {
	count := c.cartRepository.GetCountItems()

	metrics.SetCartItemCount(count)

	return nil
}
