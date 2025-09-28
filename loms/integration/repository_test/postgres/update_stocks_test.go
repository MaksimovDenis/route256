//go:build integration
// +build integration

package repository_test

import (
	"context"
	"route256/loms/internal/domain"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

func (s *Suite) TestUpdateStocks_Success(t provider.T) {
	t.Parallel()

	t.Title("Successful update of product stocks")

	var (
		testStock = []domain.Stock{
			{
				TotalCount: 200,
				Reserved:   50,
			},
			{
				TotalCount: 300,
				Reserved:   100,
			},
		}

		testData = make(map[domain.Sku]domain.Stock)
	)

	testData[s.testData.testSku2] = testStock[0]
	testData[s.testData.testSku3] = testStock[1]

	t.WithNewStep("get stock by sku", func(sCtx provider.StepCtx) {
		data := make([]domain.Item, 2)
		data[0] = domain.Item{
			Sku: s.testData.testSku2,
		}

		data[1] = domain.Item{
			Sku: s.testData.testSku3,
		}

		stock, err := s.stockRepo.GetStocksBySkuForUpdate(s.ctx, data)
		sCtx.Require().NoError(err)

		sCtx.Require().Equal(stock[s.testData.testSku2].Reserved, s.testData.testReserved2)
		sCtx.Require().Equal(stock[s.testData.testSku2].TotalCount, s.testData.testTotalCount2)

		sCtx.Require().Equal(stock[s.testData.testSku3].Reserved, s.testData.testReserved3)
		sCtx.Require().Equal(stock[s.testData.testSku3].TotalCount, s.testData.testTotalCount3)
	})

	t.WithNewStep("update stocks by sku", func(sCtx provider.StepCtx) {
		err := s.txManger.ReadCommitted(s.ctx, func(txCtx context.Context) error {
			err := s.stockRepo.UpdateStocks(txCtx, testData)
			sCtx.Require().NoError(err)

			return nil
		})

		sCtx.Require().NoError(err)
	})

	t.WithNewStep("get stock by sku", func(sCtx provider.StepCtx) {
		data := make([]domain.Item, 2)
		data[0] = domain.Item{
			Sku: s.testData.testSku2,
		}

		data[1] = domain.Item{
			Sku: s.testData.testSku3,
		}

		stock, err := s.stockRepo.GetStocksBySkuForUpdate(s.ctx, data)
		sCtx.Require().NoError(err)

		sCtx.Require().Equal(stock[s.testData.testSku2].Reserved, testStock[0].Reserved)
		sCtx.Require().Equal(stock[s.testData.testSku2].TotalCount, testStock[0].TotalCount)

		sCtx.Require().Equal(stock[s.testData.testSku3].Reserved, testStock[1].Reserved)
		sCtx.Require().Equal(stock[s.testData.testSku3].TotalCount, testStock[1].TotalCount)
	})
}

func (s *Suite) TestUpdateStocks_IncorrectCount(t provider.T) {
	t.Parallel()

	t.Title("Update with incorrect number of stocks Reserved > TotalCount")

	var (
		testStock = domain.Stock{
			TotalCount: 200,
			Reserved:   250,
		}

		testData = make(map[domain.Sku]domain.Stock)
	)

	testData[s.testData.testSku2] = testStock

	t.WithNewStep("update stocks by sku", func(sCtx provider.StepCtx) {
		err := s.txManger.ReadCommitted(s.ctx, func(txCtx context.Context) error {
			err := s.stockRepo.UpdateStocks(txCtx, testData)
			sCtx.Require().Error(err)

			return nil
		})

		sCtx.Require().Error(err)
	})
}

func (s *Suite) TestUpdateStocks_NonTx(t provider.T) {
	t.Parallel()

	t.Title("Update not in transaction")

	var (
		testStock = domain.Stock{
			TotalCount: 200,
			Reserved:   250,
		}

		testData = make(map[domain.Sku]domain.Stock)
	)

	testData[s.testData.testSku2] = testStock

	t.WithNewStep("update stocks by sku", func(sCtx provider.StepCtx) {
		err := s.stockRepo.UpdateStocks(s.ctx, testData)
		sCtx.Require().Error(err)
	})
}
