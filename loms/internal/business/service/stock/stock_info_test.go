package stock_test

import (
	"context"
	"route256/loms/internal/domain"
	testhelpers "route256/loms/internal/tool"
	"testing"

	"github.com/gojuno/minimock/v3"
)

func TestStocksInfo(t *testing.T) {
	t.Parallel()

	testSku := domain.Sku(12345)

	type mocks struct {
		mockGetStockBySku testhelpers.NeedCallWithErrAndResult[domain.Stock]
	}

	testCases := []struct {
		name          string
		mocks         mocks
		expectedSku   domain.Sku
		expectedStock int64
		expectedErr   error
	}{
		{
			name: "success: enough stock",
			mocks: mocks{
				mockGetStockBySku: testhelpers.NewNeedCallWithErrAndResult(domain.Stock{
					TotalCount: 20,
					Reserved:   5,
				}, nil),
			},
			expectedSku:   testSku,
			expectedStock: 15,
			expectedErr:   nil,
		},
		{
			name: "fail: repository error",
			mocks: mocks{
				mockGetStockBySku: testhelpers.NewNeedCallWithErrAndResult(domain.Stock{}, testhelpers.ErrForTest),
			},
			expectedSku:   testSku,
			expectedStock: 0,
			expectedErr:   testhelpers.ErrForTest,
		},
		{
			name: "fail: not enough stock (zero remainder)",
			mocks: mocks{
				mockGetStockBySku: testhelpers.NewNeedCallWithErrAndResult(domain.Stock{
					TotalCount: 10,
					Reserved:   10,
				}, nil),
			},
			expectedSku:   testSku,
			expectedStock: 0,
			expectedErr:   domain.ErrNotEnoughStock,
		},
		{
			name: "fail: not enough stock (negative remainder)",
			mocks: mocks{
				mockGetStockBySku: testhelpers.NewNeedCallWithErrAndResult(domain.Stock{
					TotalCount: 5,
					Reserved:   10,
				}, nil),
			},
			expectedSku:   testSku,
			expectedStock: 0,
			expectedErr:   domain.ErrNotEnoughStock,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			f := setUp(t)

			if tc.mocks.mockGetStockBySku.NeedCall {
				f.repository.GetStockBySkuMock.
					Expect(minimock.AnyContext, tc.expectedSku).
					Return(tc.mocks.mockGetStockBySku.Result, tc.mocks.mockGetStockBySku.Err)
			}

			stock, err := f.executor.StocksInfo(ctx, tc.expectedSku)

			if tc.expectedErr != nil {
				f.Error(err)
				f.ErrorContains(err, tc.expectedErr.Error())
				f.Equal(int64(0), stock)
			} else {
				f.NoError(err)
				f.Equal(tc.expectedStock, stock)
			}
		})
	}
}
