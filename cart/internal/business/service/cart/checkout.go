package cart

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

func (cs *Service) Checkout(ctx context.Context, userID uint64) (int64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartService.Checkout")
	defer span.Finish()

	cart, err := cs.GetItemsByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("cartService.GetItemsByUserID: %w", err)
	}

	orderID, err := cs.lomsClient.OrderCreate(ctx, userID, cart.Items)
	if err != nil {
		return 0, fmt.Errorf("lomsClient.OrderCreate: %w", err)
	}

	if err := cs.repository.DeleteItemsByUserID(ctx, userID); err != nil {
		return 0, fmt.Errorf("repository.DeleteItemsByUserID: %w", err)
	}

	return orderID, nil
}
