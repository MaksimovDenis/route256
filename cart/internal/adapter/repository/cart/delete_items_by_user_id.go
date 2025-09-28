package cart

import (
	"context"
	"route256/cart/internal/infra/logger"
	"route256/cart/internal/infra/metrics"
	"time"

	"github.com/opentracing/opentracing-go"
)

func (r *Repository) DeleteItemsByUserID(ctx context.Context, userID uint64) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartRepository.DeleteItemsByUserID")
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

	delete(r.cartByUserID, userID)

	logger.Infof(ctx, "Deleted all items from user's cart %v", userID)

	return nil
}
