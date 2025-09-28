package utils

import (
	"route256/loms/internal/business/tool/converter"
	"route256/loms/internal/domain"
	desc "route256/loms/internal/pb/loms/v1"
)

func MapItemsToDomain(items []*desc.Item) []domain.Item {
	result := make([]domain.Item, len(items))

	for idx, value := range items {
		result[idx] = domain.Item{
			Sku:   domain.Sku(value.Sku),
			Count: int64(value.Count),
		}
	}

	return result
}

func ItemsDomainToMap(items []domain.Item) ([]*desc.Item, error) {
	result := make([]*desc.Item, len(items))

	for idx, value := range items {
		checkedVal, err := converter.SafeInt64ToUint32(value.Count)
		if err != nil {
			return nil, err
		}

		result[idx] = &desc.Item{
			Sku:   int64(value.Sku),
			Count: checkedVal,
		}
	}

	return result, nil
}
