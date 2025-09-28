package cart

import (
	"context"
	"route256/cart/internal/domain"
	"route256/cart/internal/infra/logger"
	"route256/cart/internal/infra/metrics"
	"time"

	"github.com/opentracing/opentracing-go"
)

func (r *Repository) AddItem(ctx context.Context, userID uint64, item domain.Item) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartRepository.AddItem")
	defer func(now time.Time) {
		status := string(metrics.StorageQueryStatusOK)
		if err != nil {
			status = string(metrics.StorageStatusError)
		}

		metrics.IncStorageQueryCounter(string(metrics.Create), status)
		metrics.StorageQueryDurationHistogram(string(metrics.Create), status, time.Since(now).Seconds())

		span.Finish()
	}(time.Now())

	r.mx.Lock()
	defer r.mx.Unlock()

	if _, ok := r.cartByUserID[userID]; !ok {
		r.cartByUserID[userID] = make(map[domain.Sku]domain.Item)
	}

	if existing, ok := r.cartByUserID[userID][item.Sku]; ok {
		existing.Count += item.Count
		r.cartByUserID[userID][item.Sku] = existing

		logger.Infof(ctx, "Updated item count in cart for userID %v", userID)

	} else {
		copyCart := item
		r.cartByUserID[userID][item.Sku] = copyCart

		logger.Infof(ctx, "Added new item %v to cart for userID %v", item.Sku, userID)
	}

	return nil
}
