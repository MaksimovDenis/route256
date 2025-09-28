//go:build integration
// +build integration

package repository_test

import (
	"route256/loms/internal/domain"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

func (s *Suite) TestGetStockBySku_Success(t provider.T) {
	t.Parallel()

	t.Title("Successful get of stock goods")

	t.WithNewStep("get stock by sku", func(sCtx provider.StepCtx) {
		stock, err := s.stockRepo.GetStockBySku(s.ctx, s.testData.testSku1)
		sCtx.Require().NoError(err)

		sCtx.Require().Equal(stock.Reserved, s.testData.testReserved1)
		sCtx.Require().Equal(stock.TotalCount, s.testData.testTotalCount1)
	})
}

func (s *Suite) TestGetStockBySku_NotExistingOrder(t provider.T) {
	t.Parallel()

	t.Title("Receiving a non-existent stock")

	var (
		testSku = domain.Sku(1)
	)

	t.WithNewStep("get stock by sku", func(sCtx provider.StepCtx) {
		_, err := s.stockRepo.GetStockBySku(s.ctx, testSku)
		sCtx.Require().ErrorIs(err, domain.ErrStockNotFound)
	})
}
