package cart

import (
	"context"
	"fmt"
	"route256/cart/internal/domain"

	"github.com/opentracing/opentracing-go"
)

func (cs *Service) DeleteItem(ctx context.Context, userID uint64, sku domain.Sku) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartService.DeleteItem")
	defer span.Finish()

	if err := cs.repository.DeleteItem(ctx, userID, sku); err != nil {
		return fmt.Errorf("repository.DeleteItem: %w", err)
	}
	return nil
}
