package cart

import (
	"context"
	"route256/cart/internal/domain"
	"route256/cart/internal/infra/logger"
	"route256/cart/internal/infra/metrics"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestDeleteItemsByUserID(t *testing.T) {
	t.Parallel()

	err := metrics.Init(context.Background())
	require.NoError(t, err)

	err = logger.Init(zapcore.DebugLevel)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      uint64
		addItems    []*domain.Item
		expectedLen int
		expectedErr error
	}{
		{
			name:   "fail: repository.DeleteItemsByUserID ErrEmptyCart for deleted userID",
			userID: 1,
			addItems: []*domain.Item{
				{
					Sku:   domain.Sku(456),
					Count: 2,
				},
				{
					Sku:   domain.Sku(789),
					Count: 1,
				},
			},
			expectedLen: 0,
			expectedErr: domain.ErrEmptyCart,
		},
		{
			name:        "fail: repository.DeleteItemsByUserID ErrEmptyCart for non-existing userID",
			userID:      2,
			addItems:    nil,
			expectedLen: 0,
			expectedErr: domain.ErrEmptyCart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := New(10)

			for _, item := range tt.addItems {
				err := repo.AddItem(context.Background(), tt.userID, *item)
				require.NoError(t, err)
			}

			err := repo.DeleteItemsByUserID(context.Background(), tt.userID)
			require.NoError(t, err)

			items, err := repo.GetItemsByUserID(context.Background(), tt.userID)
			require.ErrorIs(t, err, tt.expectedErr)

			assert.Len(t, items, tt.expectedLen)
		})
	}
}
