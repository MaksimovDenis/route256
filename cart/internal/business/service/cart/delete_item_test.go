package cart_test

import (
	"context"
	"route256/cart/internal/domain"
	testhelpers "route256/cart/internal/tool"
	"testing"

	"github.com/gojuno/minimock/v3"
)

func TestDeleteItem(t *testing.T) {
	t.Parallel()

	testUserID := uint64(1)
	testSku := domain.Sku(12345)

	type mocks struct {
		mockDeleteItem testhelpers.NeedCallWithErr
	}

	type args struct {
		userID uint64
		sku    domain.Sku
	}

	testCases := []struct {
		name        string
		mocks       mocks
		args        args
		expectedErr error
	}{
		{
			name: "success: cartservice.DeleteItem",
			args: args{
				userID: testUserID,
				sku:    testSku,
			},
			mocks: mocks{
				mockDeleteItem: testhelpers.NewNeedCallWithErr(nil),
			},
		},
		{
			name: "fail: cartservice.DeleteItem Filed to delete item",
			args: args{
				userID: testUserID,
				sku:    testSku,
			},
			mocks: mocks{
				mockDeleteItem: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			expectedErr: testhelpers.ErrForTest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := setUp(t)

			if tc.mocks.mockDeleteItem.NeedCall {
				f.cartRepo.DeleteItemMock.
					Expect(minimock.AnyContext, tc.args.userID, tc.args.sku).
					Return(tc.mocks.mockDeleteItem.Err)
			}

			err := f.executor.DeleteItem(context.Background(), tc.args.userID, tc.args.sku)

			if tc.expectedErr != nil {
				f.ErrorIs(err, tc.expectedErr)
			} else {
				f.NoError(err)
			}
		})
	}
}
