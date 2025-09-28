package cart

import (
	"context"
	"route256/cart/internal/domain"
	"route256/cart/internal/infra/metrics"
	"time"

	"github.com/opentracing/opentracing-go"
)

func (r *Repository) GetItemOfUserIDBySku(ctx context.Context,
	userID uint64, sku domain.Sku) (item domain.Item, err error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "cartRepository.GetItemOfUserIDBySku")
	defer func(now time.Time) {
		status := string(metrics.StorageQueryStatusOK)
		if err != nil {
			status = string(metrics.StorageStatusError)
		}

		metrics.IncStorageQueryCounter(string(metrics.Select), status)
		metrics.StorageQueryDurationHistogram(string(metrics.Select), status, time.Since(now).Seconds())

		span.Finish()
	}(time.Now())

	r.mx.RLock()
	defer r.mx.RUnlock()

	userItems, ok := r.cartByUserID[userID]
	if !ok {
		return domain.Item{}, domain.ErrItemNotFound
	}

	item, ok = userItems[sku]
	if !ok {
		return domain.Item{}, domain.ErrItemNotFound
	}

	return item, nil
}
