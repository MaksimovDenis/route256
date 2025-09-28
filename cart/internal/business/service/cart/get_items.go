package cart

import (
	"context"
	"errors"
	"fmt"
	"route256/cart/internal/domain"
	"route256/cart/internal/infra/errgroup"
	"route256/cart/internal/infra/logger"
	"sort"
	"sync"

	"github.com/opentracing/opentracing-go"
)

func (cs *Service) GetItemsByUserID(ctx context.Context, userID uint64) (domain.Cart, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartService.GetItemsByUserID")
	defer span.Finish()

	items, err := cs.repository.GetItemsByUserID(ctx, userID)
	if err != nil {
		return domain.Cart{}, fmt.Errorf("repository.GetItemsByUserID: %w", err)
	}

	var (
		mx   sync.Mutex
		resp domain.Cart
	)

	group, ctx := errgroup.New(ctx)
	group.SetLimit(cs.workersCount)

	for _, item := range items {
		group.Go(func() error {
			product, err := cs.productClient.GetProductBySku(ctx, item.Sku)

			if errors.Is(err, domain.ErrProductNotFound) {
				logger.Infof(ctx, "product not found with sku %v", item.Sku)
				return nil
			}
			if err != nil {
				return fmt.Errorf("productClient.GetProductBySku: %w", err)
			}

			func() {
				defer mx.Unlock()
				mx.Lock()

				resp.Items = append(resp.Items, domain.CartItem{
					Item:    item,
					Product: product,
				})
			}()

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return domain.Cart{}, err
	}

	if len(resp.Items) == 0 {
		return domain.Cart{}, domain.ErrEmptyCart
	}

	for _, item := range resp.Items {
		resp.TotalPrice += item.Price * item.Count
	}

	sort.Slice(resp.Items, func(i, j int) bool {
		return resp.Items[i].Item.Sku < resp.Items[j].Item.Sku
	})

	return resp, nil
}
