package cart_test

import (
	"context"
	"route256/cart/internal/domain"
	testhelpers "route256/cart/internal/tool"
	"testing"

	"github.com/gojuno/minimock/v3"
)

func TestAddItem(t *testing.T) {
	t.Parallel()

	var (
		testSku    = domain.Sku(100)
		testUserID = uint64(1)
		testCount  = uint32(3)
		testStock  = uint32(5)

		testItem = domain.Item{
			Sku:   testSku,
			Count: testCount,
		}

		testProduct = domain.Product{
			Name:  "Test Product",
			Price: 1000,
			Sku:   testSku,
		}
	)

	type mocks struct {
		mockGetProductBySku      testhelpers.NeedCallWithErr
		mockStocksInfo           testhelpers.NeedCallWithErr
		mockAddItem              testhelpers.NeedCallWithErr
		mockGetItemOfUserIDBySku testhelpers.NeedCallWithErr
	}

	type args struct {
		userID             uint64
		sku                domain.Sku
		item               domain.Item
		existingStockValue uint32
		existingItem       domain.Item
		productResult      domain.Product
	}

	testCases := []struct {
		name        string
		mocks       mocks
		args        args
		expectedErr error
	}{
		{
			name: "success: cartservice.AddItem",
			mocks: mocks{
				mockGetProductBySku:      testhelpers.NewNeedCallWithErr(nil),
				mockStocksInfo:           testhelpers.NewNeedCallWithErr(nil),
				mockAddItem:              testhelpers.NewNeedCallWithErr(nil),
				mockGetItemOfUserIDBySku: testhelpers.NewNeedCallWithErr(nil),
			},
			args: args{
				userID:             testUserID,
				sku:                testSku,
				item:               testItem,
				existingStockValue: testStock,
				existingItem:       domain.Item{Sku: testSku, Count: 0},
				productResult:      testProduct,
			},
		},
		{
			name: "success: item not found in cart add new item",
			mocks: mocks{
				mockGetProductBySku:      testhelpers.NewNeedCallWithErr(nil),
				mockStocksInfo:           testhelpers.NewNeedCallWithErr(nil),
				mockAddItem:              testhelpers.NewNeedCallWithErr(nil),
				mockGetItemOfUserIDBySku: testhelpers.NewNeedCallWithErr(domain.ErrItemNotFound),
			},
			args: args{
				userID:             testUserID,
				sku:                testSku,
				item:               testItem,
				existingStockValue: testStock,
				productResult:      testProduct,
				existingItem:       domain.Item{},
			},
			expectedErr: nil,
		},
		{
			name: "fail: cartservice.AddItem ProductServiceError",
			mocks: mocks{
				mockGetProductBySku: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			args: args{
				userID:        testUserID,
				sku:           testSku,
				item:          testItem,
				productResult: domain.Product{},
			},
			expectedErr: testhelpers.ErrForTest,
		},
		{
			name: "fail: Not enough stock",
			mocks: mocks{
				mockGetProductBySku:      testhelpers.NewNeedCallWithErr(nil),
				mockStocksInfo:           testhelpers.NewNeedCallWithErr(nil),
				mockGetItemOfUserIDBySku: testhelpers.NewNeedCallWithErr(nil),
			},
			args: args{
				userID:             testUserID,
				sku:                testSku,
				item:               testItem,
				existingStockValue: 1,
				productResult:      testProduct,
				existingItem:       domain.Item{Sku: testSku, Count: 0},
			},
			expectedErr: domain.ErrNotEnoughStocks,
		},
		{
			name: "fail: AddItem repository error",
			mocks: mocks{
				mockGetProductBySku:      testhelpers.NewNeedCallWithErr(nil),
				mockStocksInfo:           testhelpers.NewNeedCallWithErr(nil),
				mockAddItem:              testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
				mockGetItemOfUserIDBySku: testhelpers.NewNeedCallWithErr(nil),
			},
			args: args{
				userID:             testUserID,
				sku:                testSku,
				item:               testItem,
				existingStockValue: testStock,
				productResult:      testProduct,
				existingItem:       domain.Item{Sku: testSku, Count: 0},
			},
			expectedErr: testhelpers.ErrForTest,
		},
		{
			name: "fail: unexpected error from GetItemOfUserIDBySku",
			mocks: mocks{
				mockGetProductBySku:      testhelpers.NewNeedCallWithErr(nil),
				mockStocksInfo:           testhelpers.NewNeedCallWithErr(nil),
				mockAddItem:              testhelpers.NeedCallWithErr{NeedCall: false},
				mockGetItemOfUserIDBySku: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			args: args{
				userID:             testUserID,
				sku:                testSku,
				item:               testItem,
				existingStockValue: testStock,
				productResult:      testProduct,
			},
			expectedErr: testhelpers.ErrForTest,
		},
		{
			name: "fail: StockInfo error",
			mocks: mocks{
				mockGetProductBySku: testhelpers.NewNeedCallWithErr(nil),
				mockStocksInfo:      testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			args: args{
				userID:             testUserID,
				sku:                testSku,
				item:               testItem,
				existingStockValue: testStock,
				productResult:      testProduct,
			},
			expectedErr: testhelpers.ErrForTest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := setUp(t)

			if tc.mocks.mockGetProductBySku.NeedCall {
				f.productClient.GetProductBySkuMock.
					Expect(minimock.AnyContext, tc.args.sku).
					Return(tc.args.productResult, tc.mocks.mockGetProductBySku.Err)
			}

			if tc.mocks.mockStocksInfo.NeedCall {
				f.lomsClient.StocksInfoMock.
					Expect(minimock.AnyContext, uint64(tc.args.item.Sku)).
					Return(int64(tc.args.existingStockValue), tc.mocks.mockStocksInfo.Err)
			}

			if tc.mocks.mockGetItemOfUserIDBySku.NeedCall {
				f.cartRepo.GetItemOfUserIDBySkuMock.
					Expect(minimock.AnyContext, tc.args.userID, tc.args.sku).
					Return(tc.args.existingItem, tc.mocks.mockGetItemOfUserIDBySku.Err)
			}

			if tc.mocks.mockAddItem.NeedCall {
				f.cartRepo.AddItemMock.
					Expect(minimock.AnyContext, tc.args.userID, testItem).
					Return(tc.mocks.mockAddItem.Err)
			}

			err := f.executor.AddItem(context.Background(), tc.args.userID, tc.args.item)

			if tc.expectedErr != nil {
				f.Error(err)
				f.ErrorContains(err, tc.expectedErr.Error())
			} else {
				f.NoError(err)
			}
		})
	}
}
