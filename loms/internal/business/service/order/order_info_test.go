package order_test

import (
	"context"
	"route256/loms/internal/domain"
	testhelpers "route256/loms/internal/tool"
	"testing"

	"github.com/gojuno/minimock/v3"
)

func TestOrderInfo(t *testing.T) {
	t.Parallel()

	testOrderID := int64(12345)

	unsortedItems := []domain.Item{
		{Sku: 3002, Count: 1},
		{Sku: 1001, Count: 2},
		{Sku: 2000, Count: 5},
	}

	sortedItems := []domain.Item{
		{Sku: 1001, Count: 2},
		{Sku: 2000, Count: 5},
		{Sku: 3002, Count: 1},
	}

	testOrder := domain.Order{
		UserID: 1,
		Status: domain.OrderStatusNew,
		Items:  unsortedItems,
	}

	type mocks struct {
		mockGetByOrderID testhelpers.NeedCallWithErr
	}

	testCases := []struct {
		name          string
		mocks         mocks
		expectedErr   error
		expectedOrder domain.Order
	}{
		{
			name: "success: orderservice.OrderInfo returns sorted items",
			mocks: mocks{
				mockGetByOrderID: testhelpers.NewNeedCallWithErr(nil),
			},
			expectedErr: nil,
			expectedOrder: domain.Order{
				UserID: testOrder.UserID,
				Status: testOrder.Status,
				Items:  sortedItems,
			},
		},
		{
			name: "fail: orderservice.OrderInfo GetByOrderID error",
			mocks: mocks{
				mockGetByOrderID: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			expectedErr:   testhelpers.ErrForTest,
			expectedOrder: domain.Order{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			f := setUp(t)

			if tc.mocks.mockGetByOrderID.NeedCall {
				f.orderRepository.GetByOrderIDMock.
					Expect(minimock.AnyContext, testOrderID).
					Return(testOrder, tc.mocks.mockGetByOrderID.Err)
			}

			order, err := f.executor.OrderInfo(ctx, testOrderID)

			if tc.expectedErr != nil {
				f.Error(err)
				f.ErrorContains(err, tc.expectedErr.Error())
				f.Equal(domain.Order{}, order)
			} else {
				f.NoError(err)
				f.Equal(tc.expectedOrder.UserID, order.UserID)
				f.Equal(tc.expectedOrder.Status, order.Status)
				f.Equal(tc.expectedOrder.Items, order.Items)
			}
		})
	}
}
