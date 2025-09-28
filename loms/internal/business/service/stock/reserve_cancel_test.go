package stock_test

import (
	"context"
	"fmt"
	"route256/loms/internal/domain"
	testhelpers "route256/loms/internal/tool"
	"testing"

	"github.com/gojuno/minimock/v3"
)

func TestReserveCancel(t *testing.T) {
	t.Parallel()

	testItem := domain.Item{Sku: 1001, Count: 2}
	expected := make(map[domain.Sku]domain.Stock)
	expected[1001] = domain.Stock{
		TotalCount: 10,
		Reserved:   3,
	}

	type mocks struct {
		getStockBySkuForUpdate testhelpers.NeedCallWithErrAndResult[map[domain.Sku]domain.Stock]
		updateStocks           testhelpers.NeedCallWithErr
	}

	testCases := []struct {
		name        string
		mocks       mocks
		expected    map[domain.Sku]domain.Stock
		expectedErr error
	}{
		{
			name: "success: reserve cancel successful",
			mocks: mocks{
				getStockBySkuForUpdate: testhelpers.NewNeedCallWithErrAndResult(map[domain.Sku]domain.Stock{
					1001: {
						TotalCount: 10,
						Reserved:   5,
					},
				}, nil),
				updateStocks: testhelpers.NewNeedCallWithErr(nil),
			},
			expected:    expected,
			expectedErr: nil,
		},
		{
			name: "fail: GetStockBySkuForUpdate returns error",
			mocks: mocks{
				getStockBySkuForUpdate: testhelpers.NewNeedCallWithErrAndResult(map[domain.Sku]domain.Stock{}, testhelpers.ErrForTest),
			},
			expectedErr: fmt.Errorf("stockRepository.GetStockBySkuForUpdate: %w", testhelpers.ErrForTest),
		},
		{
			name: "fail: not enough reserved",
			mocks: mocks{
				getStockBySkuForUpdate: testhelpers.NewNeedCallWithErrAndResult(map[domain.Sku]domain.Stock{
					1001: {
						TotalCount: 10,
						Reserved:   1,
					},
				}, nil),
			},
			expectedErr: domain.ErrInvalidReserveOperation,
		},
		{
			name: "fail: UpdateStocks returns error",
			mocks: mocks{
				getStockBySkuForUpdate: testhelpers.NewNeedCallWithErrAndResult(map[domain.Sku]domain.Stock{
					1001: {
						TotalCount: 10,
						Reserved:   5,
					},
				}, nil),
				updateStocks: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			expected:    expected,
			expectedErr: fmt.Errorf("stockRepository.UpdateStockCount: %w", testhelpers.ErrForTest),
		},
		{
			name: "fail: sku not found in returned map",
			mocks: mocks{
				getStockBySkuForUpdate: testhelpers.NewNeedCallWithErrAndResult(map[domain.Sku]domain.Stock{}, nil),
			},
			expectedErr: fmt.Errorf("%w: sku %v", domain.ErrStockNotFound, testItem.Sku),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			f := setUp(t)

			if tc.mocks.getStockBySkuForUpdate.NeedCall {
				f.repository.GetStocksBySkuForUpdateMock.
					Expect(minimock.AnyContext, []domain.Item{testItem}).
					Return(tc.mocks.getStockBySkuForUpdate.Result, tc.mocks.getStockBySkuForUpdate.Err)
			}

			if tc.mocks.updateStocks.NeedCall {
				f.repository.UpdateStocksMock.
					Expect(minimock.AnyContext, tc.expected).
					Return(tc.mocks.updateStocks.Err)
			}

			err := f.executor.ReserveCancel(ctx, []domain.Item{testItem})

			if tc.expectedErr != nil {
				f.Error(err)
				f.ErrorContains(err, tc.expectedErr.Error())
			} else {
				f.NoError(err)
			}
		})
	}
}
