package cart

import (
	"context"
	"route256/cart/internal/domain"
	"route256/cart/internal/infra/logger"
	"route256/cart/internal/infra/metrics"
	"time"

	"github.com/opentracing/opentracing-go"
)

func (r *Repository) DeleteItem(ctx context.Context, userID uint64, sku domain.Sku) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartRepository.DeleteItem")
	defer func(now time.Time) {
		status := string(metrics.StorageQueryStatusOK)
		if err != nil {
			status = string(metrics.StorageStatusError)
		}

		metrics.IncStorageQueryCounter(string(metrics.Delete), status)
		metrics.StorageQueryDurationHistogram(string(metrics.Delete), status, time.Since(now).Seconds())

		span.Finish()
	}(time.Now())

	r.mx.Lock()
	defer r.mx.Unlock()

	if _, ok := r.cartByUserID[userID]; !ok {
		logger.Warnf(ctx, "Attempted to delete item, but no such cart found %v", userID)

		return nil
	}

	delete(r.cartByUserID[userID], sku)

	logger.Infof(ctx, "Deleted item %v from cart %v", sku, userID)

	if len(r.cartByUserID[userID]) == 0 {
		delete(r.cartByUserID, userID)

		logger.Infof(ctx, "Cart for userID %v is now empty, deleted from storage", userID)
	}

	return nil
}
