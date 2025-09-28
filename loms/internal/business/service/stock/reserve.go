package stock

import (
	"context"
	"fmt"
	"route256/loms/internal/domain"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) Reserve(ctx context.Context, items []domain.Item) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "stockService.Reserve")
	defer span.Finish()

	stocks, err := s.stockRepository.GetStocksBySkuForUpdate(ctx, items)
	if err != nil {
		return fmt.Errorf("stockRepository.GetStockBySkuForUpdate: %w", err)
	}

	for _, item := range items {
		stock, ok := stocks[item.Sku]
		if !ok {
			return fmt.Errorf("%w: sku %v", domain.ErrStockNotFound, item.Sku)
		}

		if stock.TotalCount < (stock.Reserved + item.Count) {
			return domain.ErrNotEnoughStock
		}

		stock.Reserved += item.Count

		stocks[item.Sku] = stock
	}

	if err := s.stockRepository.UpdateStocks(ctx, stocks); err != nil {
		return fmt.Errorf("stockRepository.UpdateStockCount: %w", err)
	}

	return nil
}
