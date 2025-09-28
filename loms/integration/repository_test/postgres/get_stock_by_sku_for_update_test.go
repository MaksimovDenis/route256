//go:build integration
// +build integration

package repository_test

import (
	"route256/loms/internal/domain"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

func (s *Suite) TestGetStocksBySkuForUpate_Success(t provider.T) {
	t.Parallel()

	t.Title("Successful get of stock goods for update")

	t.WithNewStep("get stock by sku", func(sCtx provider.StepCtx) {
		data := make([]domain.Item, 1)
		data[0] = domain.Item{
			Sku: s.testData.testSku1,
		}

		stock, err := s.stockRepo.GetStocksBySkuForUpdate(s.ctx, data)
		sCtx.Require().NoError(err)

		sCtx.Require().Equal(stock[s.testData.testSku1].Reserved, s.testData.testReserved1)
		sCtx.Require().Equal(stock[s.testData.testSku1].TotalCount, s.testData.testTotalCount1)
	})
}

func (s *Suite) TestGetStocksBySkuForUpate_NotExistingOrde(t provider.T) {
	t.Parallel()

	t.Title("Receiving a non-existent stock")

	t.WithNewStep("get stock by sku", func(sCtx provider.StepCtx) {
		data := make([]domain.Item, 1)

		data[0] = domain.Item{
			Sku: 1,
		}

		_, err := s.stockRepo.GetStocksBySkuForUpdate(s.ctx, data)
		sCtx.Require().ErrorIs(err, domain.ErrStockNotFound)
	})
}
