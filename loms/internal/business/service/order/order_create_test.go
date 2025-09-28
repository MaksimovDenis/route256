package order_test

import (
	"context"
	"errors"
	"fmt"
	"route256/loms/internal/domain"
	txmanager "route256/loms/internal/infra/tx_manager"
	testhelpers "route256/loms/internal/tool"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
)

func TestOrderCreate(t *testing.T) {
	t.Parallel()

	testOrderID := int64(12345)
	testItems := []domain.Item{{Sku: 1001, Count: 2}}
	testOrder := domain.Order{UserID: 1, Items: testItems}

	type mocks struct {
		createOrder             testhelpers.NeedCallWithErr
		createOrderItems        testhelpers.NeedCallWithErr
		reserve                 testhelpers.NeedCallWithErr
		setStatusAndCreateEvent testhelpers.NeedCallWithErr
		sentEvent               testhelpers.NeedCallWithErr
	}

	testCases := []struct {
		name           string
		mocks          mocks
		expectedErr    string
		expectedID     int64
		expectedStatus domain.EventStatus
	}{
		{
			name: "success: order creation and reserve successful",
			mocks: mocks{
				createOrder:             testhelpers.NewNeedCallWithErr(nil),
				createOrderItems:        testhelpers.NewNeedCallWithErr(nil),
				reserve:                 testhelpers.NewNeedCallWithErr(nil),
				setStatusAndCreateEvent: testhelpers.NewNeedCallWithErr(nil),
				sentEvent:               testhelpers.NewNeedCallWithErr(nil),
			},
			expectedErr:    "",
			expectedID:     testOrderID,
			expectedStatus: domain.EventStatusNew,
		},
		{
			name: "fail: CreateOrder returns error",
			mocks: mocks{
				createOrder: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			expectedErr: "orderRepository.CreateOrder",
			expectedID:  0,
		},
		{
			name: "fail: CreateOrderItems returns error",
			mocks: mocks{
				createOrder:      testhelpers.NewNeedCallWithErr(nil),
				createOrderItems: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			expectedErr: "orderRepository.CreateOrderItems",
			expectedID:  0,
		},
		{
			name: "fail: Reserve returns domain.ErrNotEnoughStock, SetStatusAndCreateEvent fails",
			mocks: mocks{
				createOrder:             testhelpers.NewNeedCallWithErr(nil),
				createOrderItems:        testhelpers.NewNeedCallWithErr(nil),
				reserve:                 testhelpers.NewNeedCallWithErr(domain.ErrNotEnoughStock),
				setStatusAndCreateEvent: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
				sentEvent:               testhelpers.NewNeedCallWithErr(nil),
			},
			expectedErr:    "setStatusAndCreateEvent",
			expectedID:     0,
			expectedStatus: domain.EventStatusNew,
		},
		{
			name: "fail: Reserve returns unexpected error, SetStatusAndCreateEvent is NOT called",
			mocks: mocks{
				createOrder:      testhelpers.NewNeedCallWithErr(nil),
				createOrderItems: testhelpers.NewNeedCallWithErr(nil),
				reserve:          testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
				sentEvent:        testhelpers.NewNeedCallWithErr(nil),
			},
			expectedErr:    "stockService.Reserve",
			expectedID:     0,
			expectedStatus: domain.EventStatusNew,
		},
		{
			name: "fail: SetStatusAndCreateEvent AwaitingPayment returns error",
			mocks: mocks{
				createOrder:             testhelpers.NewNeedCallWithErr(nil),
				createOrderItems:        testhelpers.NewNeedCallWithErr(nil),
				reserve:                 testhelpers.NewNeedCallWithErr(nil),
				setStatusAndCreateEvent: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
				sentEvent:               testhelpers.NewNeedCallWithErr(nil),
			},
			expectedErr:    "setStatusAndCreateEvent",
			expectedID:     0,
			expectedStatus: domain.EventStatusNew,
		},
		{
			name: "fail: eventRepository.CreateEvent error",
			mocks: mocks{
				createOrder:      testhelpers.NewNeedCallWithErr(nil),
				createOrderItems: testhelpers.NewNeedCallWithErr(nil),
				sentEvent:        testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			expectedErr:    "eventRepository.CreateEvent",
			expectedID:     0,
			expectedStatus: domain.EventStatusNew,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			f := setUp(t)

			callCount := 0
			f.txManager.ReadCommittedMock.Set(func(ctx context.Context, fn txmanager.Handler) error {
				callCount++
				return fn(ctx)
			})

			if tc.mocks.createOrder.NeedCall {
				f.orderRepository.CreateOrderMock.
					Expect(minimock.AnyContext, testOrder.UserID).
					Return(testOrderID, tc.mocks.createOrder.Err)
			}

			if tc.mocks.createOrderItems.NeedCall {
				f.orderRepository.CreateOrderItemsMock.
					Expect(minimock.AnyContext, testOrderID, testItems).
					Return(tc.mocks.createOrderItems.Err)
			}

			if tc.mocks.reserve.NeedCall {
				f.stockService.ReserveMock.
					Expect(minimock.AnyContext, testOrder.Items).
					Return(tc.mocks.reserve.Err)
			}

			if tc.mocks.setStatusAndCreateEvent.NeedCall {
				f.orderRepository.SetStatusAndCreateEventMock.Set(func(_ context.Context, orderID int64, status domain.OrderStatus, _ domain.Event) error {
					if orderID != testOrderID {
						return fmt.Errorf("unexpected orderID: got %d", orderID)
					}

					expectedStatus := domain.OrderStatusAwaitingPayment
					if tc.mocks.reserve.Err != nil && errors.Is(tc.mocks.reserve.Err, domain.ErrNotEnoughStock) {
						expectedStatus = domain.OrderStatusFailed
					}
					if status != expectedStatus {
						return fmt.Errorf("unexpected status: got %v, want %v", status, expectedStatus)
					}

					return tc.mocks.setStatusAndCreateEvent.Err
				})
			}

			if tc.mocks.sentEvent.NeedCall {
				f.eventRepository.CreateEventMock.Set(func(_ context.Context, event domain.Event) error {
					require.Equal(t, tc.expectedStatus, event.Status)
					return tc.mocks.sentEvent.Err
				})
			}

			id, err := f.executor.OrderCreate(ctx, testOrder)

			if tc.expectedErr != "" {
				f.Error(err)
				f.ErrorContains(err, tc.expectedErr)
				f.Equal(tc.expectedID, id)
			} else {
				f.NoError(err)
				f.Equal(tc.expectedID, id)
			}
		})
	}
}
