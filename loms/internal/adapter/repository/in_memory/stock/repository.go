package stock

import (
	_ "embed"
	"encoding/json"
	"maps"
	"route256/loms/internal/domain"
	"sync"
)

//go:embed stock-data.json
var stockData []byte

type stockDTO struct {
	Sku        domain.Sku `json:"sku"`
	TotalCount int64      `json:"total_count"`
	Reserved   int64      `json:"reserved"`
}

type stockBySku = map[domain.Sku]domain.Stock

type Repository struct {
	stockBySku stockBySku
	mx         sync.RWMutex
}

func New() (*Repository, error) {
	var data []stockDTO
	err := json.Unmarshal(stockData, &data)
	if err != nil {
		return nil, err
	}

	stockMap := make(stockBySku, len(data))

	for _, item := range data {
		stockMap[item.Sku] = domain.Stock{
			TotalCount: item.TotalCount,
			Reserved:   item.Reserved,
		}
	}

	return &Repository{
		stockBySku: stockMap,
	}, nil
}

func (r *Repository) GetStockBySku(sku domain.Sku) (int64, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	stocks, ok := r.stockBySku[sku]
	if !ok {
		return 0, domain.ErrStockNotFound
	}

	remainderCount := stocks.TotalCount - stocks.Reserved

	if remainderCount <= 0 {
		return 0, domain.ErrNotEnoughStock
	}

	return remainderCount, nil
}

func (r *Repository) Reserve(items []domain.Item) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	updatedStocks := make(map[domain.Sku]domain.Stock)

	for _, value := range items {
		stocks, ok := r.stockBySku[value.Sku]
		if !ok {
			return domain.ErrStockNotFound
		}

		if stocks.TotalCount < (stocks.Reserved + value.Count) {
			return domain.ErrNotEnoughStock
		}

		stocks.Reserved += value.Count
		updatedStocks[value.Sku] = stocks
	}

	maps.Copy(r.stockBySku, updatedStocks)

	return nil
}

func (r *Repository) ReserveRemove(items []domain.Item) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	updatedStocks := make(map[domain.Sku]domain.Stock)

	for _, value := range items {
		stocks, ok := r.stockBySku[value.Sku]
		if !ok {
			return domain.ErrStockNotFound
		}

		if stocks.TotalCount < value.Count || stocks.Reserved < value.Count {
			return domain.ErrInvalidReserveOperation
		}

		stocks.TotalCount -= value.Count
		stocks.Reserved -= value.Count

		updatedStocks[value.Sku] = stocks
	}

	maps.Copy(r.stockBySku, updatedStocks)

	return nil
}

func (r *Repository) ReserveCancel(items []domain.Item) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	updatedStocks := make(map[domain.Sku]domain.Stock)

	for _, value := range items {
		stocks, ok := r.stockBySku[value.Sku]
		if !ok {
			return domain.ErrStockNotFound
		}

		if stocks.Reserved < value.Count {
			return domain.ErrInvalidReserveOperation
		}

		stocks.Reserved -= value.Count
		updatedStocks[value.Sku] = stocks
	}

	maps.Copy(r.stockBySku, updatedStocks)

	return nil
}
