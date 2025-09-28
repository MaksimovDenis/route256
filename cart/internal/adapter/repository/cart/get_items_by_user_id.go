package cart

import (
	"context"
	"route256/cart/internal/domain"
	"route256/cart/internal/infra/logger"
	"route256/cart/internal/infra/metrics"
	"time"

	"github.com/opentracing/opentracing-go"
)

func (r *Repository) GetItemsByUserID(ctx context.Context,
	userID uint64) (items []domain.Item, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartRepository.GetItemsByUserID")
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
		return nil, domain.ErrEmptyCart
	}

	items = make([]domain.Item, 0, len(userItems))
	for _, item := range userItems {
		items = append(items, item)
	}

	logger.Infof(ctx, "Fetched items from userID %v cart", userID)

	return items, nil
}
