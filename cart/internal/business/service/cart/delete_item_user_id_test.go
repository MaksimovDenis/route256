package cart_test

import (
	"context"
	testhelpers "route256/cart/internal/tool"
	"testing"

	"github.com/gojuno/minimock/v3"
)

func TestDeleteItemsByUserID(t *testing.T) {
	t.Parallel()

	testUserID := uint64(1)

	type mocks struct {
		mockDeleteItems testhelpers.NeedCallWithErr
	}

	testCases := []struct {
		name        string
		userID      uint64
		mocks       mocks
		expectedErr error
	}{
		{
			name:   "success: cartservice.DeleteItemsByUserID",
			userID: testUserID,
			mocks: mocks{
				mockDeleteItems: testhelpers.NewNeedCallWithErr(nil),
			},
		},
		{
			name:   "fail: cartservice.DeleteItemsByUserID Filed to delete item",
			userID: testUserID,
			mocks: mocks{
				mockDeleteItems: testhelpers.NewNeedCallWithErr(testhelpers.ErrForTest),
			},
			expectedErr: testhelpers.ErrForTest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := setUp(t)

			if tc.mocks.mockDeleteItems.NeedCall {
				f.cartRepo.DeleteItemsByUserIDMock.
					Expect(minimock.AnyContext, tc.userID).
					Return(tc.mocks.mockDeleteItems.Err)
			}

			err := f.executor.DeleteItemsByUserID(context.Background(), tc.userID)

			if tc.expectedErr != nil {
				f.ErrorIs(err, tc.expectedErr)
			} else {
				f.NoError(err)
			}
		})
	}
}
