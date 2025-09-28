package stock

import (
	"context"
	"route256/loms/internal/domain"
)

//go:generate rm -rf mock
//go:generate mkdir -p mock
//go:generate minimock -i * -o ./mock -s "_mock.go" -g
type stockRepository interface {
	GetStockBySku(ctx context.Context, sku domain.Sku) (domain.Stock, error)
	GetStocksBySkuForUpdate(ctx context.Context, items []domain.Item) (map[domain.Sku]domain.Stock, error)
	UpdateStocks(ctx context.Context, stocks map[domain.Sku]domain.Stock) error
}

type Service struct {
	stockRepository stockRepository
}

func New(stockRepository stockRepository) *Service {
	return &Service{
		stockRepository: stockRepository,
	}
}
