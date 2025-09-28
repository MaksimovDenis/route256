package cart_test

import (
	"context"
	"route256/cart/internal/domain"
	testhelpers "route256/cart/internal/tool"
	"testing"

	"github.com/gojuno/minimock/v3"
)

func TestCheckout(t *testing.T) {
	t.Parallel()

	var (
		testSku    = domain.Sku(100)
		testUserID = uint64(1)

		testItem = domain.Item{Sku: testSku, Count: 2}

		testProduct = domain.Product{
			Name:  "Test Product",
			Price: 1500,
			Sku:   testSku,
		}

		testOrderID = int64(42)
	)

	type mocks struct {
		mockGetItemsByUserID    testhelpers.NeedCallWithErr
		mockGetProductBySku     testhelpers.NeedCallWithErr
		mockOrderCreate         testhelpers.NeedCallWithErr
		mockDeleteItemsByUserID testhelpers.NeedCallWithErr
	}

	type args struct {
		userID      uint64
		testItems   []domain.Item
		testCart    []domain.CartItem
		testProduct domain.Product
		testOrderID int64
	}

	testCases := []struct {
		name            string
		mocks           mocks
		args            args
		expectedOrderID int64
		expectedErr     error
	}{
		{
			name: "success: checkout completes successfully",
			mocks: mocks{
				mockGetItemsByUserID:    testhelpers.NewNeedCallWithErr(nil),
				mockGetProductBySku:     testhelpers.NewNeedCallWithErr(nil),
				mockOrderCreate:         testhelpers.NewNeedCallWithErr(nil),
				mockDeleteItemsByUserID: testhelpers.NewNeedCallWithErr(nil),
			},
			args: args{
				userID:    testUserID,
				testItems: []domain.Item{testItem},
				testCart: []domain.CartItem{
					{
						Item:    testItem,
						Product: testProduct,
					},
				},
				testProduct: testProduct,
				testOrderID: testOrderID,
			},
			expectedOrderID: testOrderID,
		},
		{
			name: "fail: GetItemsByUserID returns error",
			mocks: mocks{
				mockGetItemsByUserID: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			args: args{
				userID: testUserID,
			},
			expectedErr: testhelpers.ErrForTest,
		},
		{
			name: "fail: OrderCreate returns error",
			mocks: mocks{
				mockGetItemsByUserID: testhelpers.NewNeedCallWithErr(nil),
				mockGetProductBySku:  testhelpers.NewNeedCallWithErr(nil),
				mockOrderCreate:      testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			args: args{
				userID:    testUserID,
				testItems: []domain.Item{testItem},
				testCart: []domain.CartItem{
					{
						Item:    testItem,
						Product: testProduct,
					},
				},
				testProduct: testProduct,
			},
			expectedErr: testhelpers.ErrForTest,
		},
		{
			name: "fail: DeleteItemsByUserID returns error",
			mocks: mocks{
				mockGetItemsByUserID:    testhelpers.NewNeedCallWithErr(nil),
				mockGetProductBySku:     testhelpers.NewNeedCallWithErr(nil),
				mockOrderCreate:         testhelpers.NewNeedCallWithErr(nil),
				mockDeleteItemsByUserID: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			args: args{
				userID:    testUserID,
				testItems: []domain.Item{testItem},
				testCart: []domain.CartItem{
					{
						Item:    testItem,
						Product: testProduct,
					},
				},
				testProduct: testProduct,
			},
			expectedErr: testhelpers.ErrForTest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := setUp(t)

			if tc.mocks.mockGetItemsByUserID.NeedCall {
				f.cartRepo.GetItemsByUserIDMock.
					Expect(minimock.AnyContext, tc.args.userID).
					Return(tc.args.testItems, tc.mocks.mockGetItemsByUserID.Err)
			}

			if tc.mocks.mockGetProductBySku.NeedCall {
				for _, item := range tc.args.testItems {
					f.productClient.GetProductBySkuMock.
						Expect(minimock.AnyContext, item.Sku).
						Return(tc.args.testProduct, tc.mocks.mockGetProductBySku.Err)
				}
			}

			if tc.mocks.mockOrderCreate.NeedCall {
				f.lomsClient.OrderCreateMock.
					Expect(minimock.AnyContext, tc.args.userID, tc.args.testCart).
					Return(tc.args.testOrderID, tc.mocks.mockOrderCreate.Err)
			}

			if tc.mocks.mockDeleteItemsByUserID.NeedCall {
				f.cartRepo.DeleteItemsByUserIDMock.
					Expect(minimock.AnyContext, tc.args.userID).
					Return(tc.mocks.mockDeleteItemsByUserID.Err)
			}

			gotOrderID, err := f.executor.Checkout(context.Background(), tc.args.userID)

			if tc.expectedErr != nil {
				f.ErrorIs(err, tc.expectedErr)
				f.Equal(int64(0), gotOrderID)
			} else {
				f.NoError(err)
				f.Equal(tc.expectedOrderID, gotOrderID)
			}
		})
	}
}
