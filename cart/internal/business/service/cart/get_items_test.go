package cart_test

import (
	"context"
	"fmt"
	"route256/cart/internal/domain"
	"route256/cart/internal/infra/logger"
	testhelpers "route256/cart/internal/tool"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestGetItemsByUserID(t *testing.T) {
	t.Parallel()

	err := logger.Init(zapcore.DebugLevel)
	require.NoError(t, err)

	var (
		testSku    = domain.Sku(100)
		testUserID = uint64(1)

		testItem = domain.Item{Sku: testSku, Count: 2}

		testProduct = domain.Product{
			Name:  "Test Product",
			Price: 1500,
			Sku:   testSku,
		}

		testCart = domain.Cart{
			Items: []domain.CartItem{
				{
					Item:    testItem,
					Product: testProduct,
				},
			},
			TotalPrice: 1500 * 2,
		}
	)

	type mocks struct {
		mockGetItemsUseByUserID testhelpers.NeedCallWithErr
		mockGetProductBySku     testhelpers.NeedCallWithErr
	}

	testCases := []struct {
		name              string
		userID            uint64
		mockItems         []domain.Item
		mocks             mocks
		mockProductResult domain.Product
		wantCart          domain.Cart
		expectedErr       error
	}{
		{
			name:      "success: cartservice.GetItemsByUserID",
			userID:    testUserID,
			mockItems: []domain.Item{testItem},
			mocks: mocks{
				mockGetItemsUseByUserID: testhelpers.NewNeedCallWithErr(nil),
				mockGetProductBySku:     testhelpers.NewNeedCallWithErr(nil),
			},
			mockProductResult: testProduct,
			wantCart:          testCart,
		},
		{
			name:   "fail: cartservice.GetItemsByUserID Repository Error",
			userID: testUserID,
			mocks: mocks{
				mockGetItemsUseByUserID: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			expectedErr: testhelpers.ErrForTest,
		},
		{
			name:      "fail: cartservice.GetItemsByUserID ErrEmptyCart",
			userID:    testUserID,
			mockItems: []domain.Item{},
			mocks: mocks{
				mockGetItemsUseByUserID: testhelpers.NewNeedCallWithErr(nil),
			},
			expectedErr: domain.ErrEmptyCart,
		},
		{
			name:      "fail: cartservice.GetItemsByUserID Product Service Error",
			userID:    testUserID,
			mockItems: []domain.Item{testItem},
			mocks: mocks{
				mockGetItemsUseByUserID: testhelpers.NewNeedCallWithErr(nil),
				mockGetProductBySku:     testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			expectedErr: testhelpers.ErrForTest,
		},
		{
			name:      "fail: cartservice.GetItemsByUserID ErrIncorrectSku",
			userID:    testUserID,
			mockItems: []domain.Item{testItem},
			mocks: mocks{
				mockGetItemsUseByUserID: testhelpers.NewNeedCallWithErr(nil),
				mockGetProductBySku:     testhelpers.NewNeedCallWithErr(domain.ErrIncorrectSku),
			},
			expectedErr: domain.ErrIncorrectSku,
		},
		{
			name:      "fail: cartservice.GetItemsByUserID ErrEmptyCart (item not found the skiped)",
			userID:    testUserID,
			mockItems: []domain.Item{testItem},
			mocks: mocks{
				mockGetItemsUseByUserID: testhelpers.NewNeedCallWithErr(nil),
				mockGetProductBySku:     testhelpers.NewNeedCallWithErr(domain.ErrProductNotFound),
			},
			expectedErr: domain.ErrEmptyCart,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := setUp(t)

			if tc.mocks.mockGetItemsUseByUserID.NeedCall {
				f.cartRepo.GetItemsByUserIDMock.
					Expect(minimock.AnyContext, tc.userID).
					Return(tc.mockItems, tc.mocks.mockGetItemsUseByUserID.Err)
			}

			if tc.mocks.mockGetProductBySku.NeedCall {
				f.productClient.GetProductBySkuMock.
					Expect(minimock.AnyContext, testSku).
					Return(tc.mockProductResult, tc.mocks.mockGetProductBySku.Err)
			}

			got, err := f.executor.GetItemsByUserID(context.Background(), tc.userID)

			if tc.expectedErr != nil {
				f.ErrorIs(err, tc.expectedErr)
			} else {
				f.NoError(err)
			}

			f.Equal(tc.wantCart, got)
		})
	}
}

func TestGetItemsByUserIDSortsBySku(t *testing.T) {
	t.Parallel()

	testUserID := uint64(1)

	mockItems := []domain.Item{
		{Sku: 300, Count: 1},
		{Sku: 100, Count: 1},
	}

	products := map[domain.Sku]domain.Product{
		100: {Name: "Product 100", Price: 1000, Sku: 100},
		300: {Name: "Product 300", Price: 2000, Sku: 300},
	}

	wantCart := domain.Cart{
		Items: []domain.CartItem{
			{
				Item:    domain.Item{Sku: 100, Count: 1},
				Product: products[100],
			},
			{
				Item:    domain.Item{Sku: 300, Count: 1},
				Product: products[300],
			},
		},
		TotalPrice: 1000 + 2000,
	}

	f := setUp(t)

	f.cartRepo.GetItemsByUserIDMock.
		Expect(minimock.AnyContext, testUserID).
		Return(mockItems, nil)

	f.productClient.GetProductBySkuMock.Set(func(_ context.Context, sku domain.Sku) (domain.Product, error) {
		if p, ok := products[sku]; ok {
			return p, nil
		}
		return domain.Product{}, fmt.Errorf("unexpected sku: %d", sku)
	})

	got, err := f.executor.GetItemsByUserID(context.Background(), testUserID)
	f.NoError(err)
	f.Equal(wantCart, got)
}
