//go:build integration
// +build integration

package repository_test

import (
	"context"
	"route256/loms/internal/domain"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

func (s *Suite) TestCreateOrder_Success(t provider.T) {
	t.Parallel()

	t.Title("Successful order creation")

	var (
		testUserID int64 = 1
		testItem         = []domain.Item{
			{
				Sku:   domain.Sku(1625903),
				Count: 10,
			},
		}
		testOrderID int64
	)

	t.WithNewStep("create order", func(sCtx provider.StepCtx) {
		var orderID int64
		var err error

		err = s.txManger.ReadCommitted(s.ctx, func(txCtx context.Context) error {
			orderID, err = s.orderRepo.CreateOrder(txCtx, testUserID)
			sCtx.Require().NoError(err)

			sCtx.Require().Greater(orderID, int64(0))

			err = s.orderRepo.CreateOrderItems(txCtx, orderID, testItem)
			sCtx.Require().NoError(err)

			return nil
		})

		sCtx.Require().NoError(err)
		testOrderID = orderID
	})

	t.WithNewStep("get order by id", func(sCtx provider.StepCtx) {
		orderID, err := s.orderRepo.GetByOrderID(s.ctx, testOrderID)
		sCtx.Require().NoError(err)

		sCtx.Require().Equal(orderID.UserID, testUserID)
		sCtx.Require().Equal(orderID.Status, domain.OrderStatusNew)
		sCtx.Require().Equal(orderID.Items, []domain.Item{
			{
				Sku:   testItem[0].Sku,
				Count: testItem[0].Count,
			},
		})
	})

	t.WithNewStep("set status", func(sCtx provider.StepCtx) {
		err := s.orderRepo.SetStatus(s.ctx, testOrderID, domain.OrderStatusAwaitingPayment)
		sCtx.Require().NoError(err)
	})

	t.WithNewStep("get order by id for update", func(sCtx provider.StepCtx) {
		orderID, err := s.orderRepo.GetByOrderIDForUpdate(s.ctx, testOrderID)
		sCtx.Require().NoError(err)

		sCtx.Require().Equal(orderID.UserID, testUserID)
		sCtx.Require().Equal(orderID.Status, domain.OrderStatusAwaitingPayment)
		sCtx.Require().Equal(orderID.Items, []domain.Item{
			{
				Sku:   testItem[0].Sku,
				Count: testItem[0].Count,
			},
		})
	})

	t.WithNewStep("set incorrect status", func(sCtx provider.StepCtx) {
		err := s.orderRepo.SetStatus(s.ctx, testOrderID, domain.OrderStatus("Incorrect status"))
		sCtx.Require().Error(err)
	})
}
