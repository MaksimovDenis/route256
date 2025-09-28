package cart

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

func (cs *Service) DeleteItemsByUserID(ctx context.Context, userID uint64) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartService.DeleteItemsByUserID")
	defer span.Finish()

	if err := cs.repository.DeleteItemsByUserID(ctx, userID); err != nil {
		return fmt.Errorf("repository.DeleteItemsByUserID: %w", err)
	}
	return nil
}
