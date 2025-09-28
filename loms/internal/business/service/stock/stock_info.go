package stock

import (
	"context"
	"fmt"
	"route256/loms/internal/domain"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) StocksInfo(ctx context.Context, sku domain.Sku) (int64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "stockService.StocksInfo")
	defer span.Finish()

	stockData, err := s.stockRepository.GetStockBySku(ctx, sku)
	if err != nil {
		return 0, fmt.Errorf("stockRepository.GetStockBySku: %w", err)
	}

	remainderCount := stockData.TotalCount - stockData.Reserved

	if remainderCount <= 0 {
		return 0, domain.ErrNotEnoughStock
	}

	return remainderCount, nil
}
