package cart

import (
	"context"
	"errors"
	"fmt"
	"route256/cart/internal/domain"

	"github.com/opentracing/opentracing-go"
)

func (cs *Service) AddItem(ctx context.Context, userID uint64, item domain.Item) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartService.AddItem")
	defer span.Finish()

	_, err := cs.productClient.GetProductBySku(ctx, item.Sku)
	if err != nil {
		return fmt.Errorf("productClient.GetProductBySku: %w", err)
	}

	count, err := cs.lomsClient.StocksInfo(ctx, uint64(item.Sku))
	if err != nil {
		return fmt.Errorf("lomsClient.StocksInfo: %w", err)
	}

	var currentCount uint32
	currentItem, err := cs.repository.GetItemOfUserIDBySku(ctx, userID, item.Sku)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			currentCount = 0
		} else {
			return fmt.Errorf("repository.GetItemOfUserIDBySku: %w", err)
		}
	} else {
		currentCount = currentItem.Count
	}

	newCount := currentCount + item.Count
	if int64(newCount) > count {
		return domain.ErrNotEnoughStocks
	}

	if err := cs.repository.AddItem(ctx, userID, item); err != nil {
		return fmt.Errorf("repository.AddItem: %w", err)
	}

	return nil
}
