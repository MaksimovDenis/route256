package order_test

import (
	"context"
	"fmt"
	"route256/loms/internal/domain"
	txmanager "route256/loms/internal/infra/tx_manager"
	testhelpers "route256/loms/internal/tool"
	"testing"

	"github.com/gojuno/minimock/v3"
)

func TestOrderPay(t *testing.T) {
	t.Parallel()

	testOrderID := int64(12345)
	testItems := []domain.Item{
		{Sku: 1001, Count: 2},
		{Sku: 1002, Count: 1},
	}

	type mocks struct {
		mockGetByOrderIDForUpdate   testhelpers.NeedCallWithErr
		mockReserveRemove           testhelpers.NeedCallWithErr
		mockReadCommitted           testhelpers.NeedCallWithErr
		mockSetStatusAndCreateEvent testhelpers.NeedCallWithErr
	}

	testCases := []struct {
		name           string
		order          domain.Order
		mocks          mocks
		expectedErr    error
		expectedStatus domain.EventStatus
	}{
		{
			name: "success: orderservice.OrderPay",
			order: domain.Order{
				UserID: 1,
				Status: domain.OrderStatusAwaitingPayment,
				Items:  testItems,
			},
			mocks: mocks{
				mockGetByOrderIDForUpdate:   testhelpers.NewNeedCallWithErr(nil),
				mockReserveRemove:           testhelpers.NewNeedCallWithErr(nil),
				mockSetStatusAndCreateEvent: testhelpers.NewNeedCallWithErr(nil),
				mockReadCommitted:           testhelpers.NewNeedCallWithErr(nil),
			},
			expectedErr:    nil,
			expectedStatus: domain.EventStatusNew,
		},
		{
			name: "success: orderservice.OrderPay already payed order returns nil",
			order: domain.Order{
				UserID: 1,
				Status: domain.OrderStatusPayed,
				Items:  testItems,
			},
			mocks: mocks{
				mockGetByOrderIDForUpdate: testhelpers.NewNeedCallWithErr(nil),
			},
			expectedErr: nil,
		},
		{
			name: "fail: orderservice.OrderPay order in wrong status",
			order: domain.Order{
				UserID: 1,
				Status: domain.OrderStatusNew,
				Items:  testItems,
			},
			mocks: mocks{
				mockGetByOrderIDForUpdate: testhelpers.NewNeedCallWithErr(nil),
			},
			expectedErr: domain.ErrPayStatusOrder,
		},
		{
			name:  "fail: orderservice.OrderPay GetByOrderID error",
			order: domain.Order{},
			mocks: mocks{
				mockGetByOrderIDForUpdate: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			expectedErr: testhelpers.ErrForTest,
		},
		{
			name: "fail: orderservice.OrderPay ReserveRemove error",
			order: domain.Order{
				UserID: 1,
				Status: domain.OrderStatusAwaitingPayment,
				Items:  testItems,
			},
			mocks: mocks{
				mockGetByOrderIDForUpdate: testhelpers.NewNeedCallWithErr(nil),
				mockReserveRemove:         testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
				mockReadCommitted:         testhelpers.NewNeedCallWithErr(nil),
			},
			expectedErr: testhelpers.ErrForTest,
		},
		{
			name: "fail: orderservice.OrderPay SetStatus error",
			order: domain.Order{
				UserID: 1,
				Status: domain.OrderStatusAwaitingPayment,
				Items:  testItems,
			},
			mocks: mocks{
				mockGetByOrderIDForUpdate:   testhelpers.NewNeedCallWithErr(nil),
				mockReserveRemove:           testhelpers.NewNeedCallWithErr(nil),
				mockSetStatusAndCreateEvent: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
				mockReadCommitted:           testhelpers.NewNeedCallWithErr(nil),
			},
			expectedErr: testhelpers.ErrForTest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := setUp(t)
			ctx := context.Background()

			f.txManager.ReadCommittedMock.Set(func(ctx context.Context, fn txmanager.Handler) error {
				return fn(ctx)
			})

			if tc.mocks.mockGetByOrderIDForUpdate.NeedCall {
				f.orderRepository.GetByOrderIDForUpdateMock.
					Expect(minimock.AnyContext, testOrderID).
					Return(tc.order, tc.mocks.mockGetByOrderIDForUpdate.Err)
			}

			if tc.mocks.mockReserveRemove.NeedCall {
				f.stockService.ReserveRemoveMock.
					Expect(minimock.AnyContext, tc.order.Items).
					Return(tc.mocks.mockReserveRemove.Err)
			}

			if tc.mocks.mockSetStatusAndCreateEvent.NeedCall {
				f.orderRepository.SetStatusAndCreateEventMock.Set(func(_ context.Context, orderID int64, status domain.OrderStatus, _ domain.Event) error {
					if orderID != testOrderID {
						return fmt.Errorf("unexpected orderID: got %d", orderID)
					}
					if status != domain.OrderStatusPayed {
						return fmt.Errorf("unexpected status: got %v", status)
					}

					return tc.mocks.mockSetStatusAndCreateEvent.Err
				})
			}

			err := f.executor.OrderPay(ctx, testOrderID)

			if tc.expectedErr != nil {
				f.Error(err)
				f.ErrorContains(err, tc.expectedErr.Error())
			} else {
				f.NoError(err)
			}
		})
	}
}
