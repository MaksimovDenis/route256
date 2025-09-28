package cart

import (
	"context"
	"route256/cart/internal/domain"
	"route256/cart/internal/infra/logger"
	"route256/cart/internal/infra/metrics"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestDeleteItem(t *testing.T) {
	t.Parallel()

	err := metrics.Init(context.Background())
	require.NoError(t, err)

	err = logger.Init(zapcore.DebugLevel)
	require.NoError(t, err)

	var (
		testSKU = []domain.Sku{1, 2}
	)

	tests := []struct {
		name        string
		userID      uint64
		addItem     []domain.Item
		deletedSku  domain.Sku
		expectedLen int
	}{
		{
			name:   "success: repository.DeleteItem",
			userID: 1,
			addItem: []domain.Item{
				{
					Sku:   testSKU[0],
					Count: 1,
				},
				{
					Sku:   testSKU[1],
					Count: 3,
				},
			},
			deletedSku:  testSKU[0],
			expectedLen: 1,
		},
		{
			name:        "success: repository.DeleteItem for non-existing item",
			userID:      2,
			addItem:     []domain.Item{},
			deletedSku:  3,
			expectedLen: 0,
		},
		{
			name:        "success: repository.DeleteItem for non-existing user",
			userID:      0,
			addItem:     []domain.Item{},
			deletedSku:  3,
			expectedLen: 0,
		},
		{
			name:        "success: repository.DeleteItem for non-existing userID",
			userID:      0,
			addItem:     []domain.Item{},
			deletedSku:  3,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := New(10)

			ctx := context.Background()

			for _, item := range tt.addItem {
				if tt.userID == 0 {
					continue
				}

				err := repo.AddItem(ctx, tt.userID, item)
				require.NoError(t, err)
			}

			err := repo.DeleteItem(ctx, tt.userID, tt.deletedSku)
			require.NoError(t, err)

			assert.Len(t, repo.cartByUserID[tt.userID], tt.expectedLen)

			err = repo.DeleteItem(ctx, tt.userID, testSKU[1])
			require.NoError(t, err)

			err = repo.DeleteItem(ctx, tt.userID, testSKU[1])
			require.NoError(t, err)
		})
	}
}

func TestDeleteItem_Concurrent(t *testing.T) {
	t.Parallel()

	err := metrics.Init(context.Background())
	require.NoError(t, err)

	err = logger.Init(zapcore.DebugLevel)
	require.NoError(t, err)

	const userID = uint64(1)

	repo := New(10)
	ctx := context.Background()

	skus := []domain.Sku{100, 101, 102, 103, 104}
	for _, sku := range skus {
		err := repo.AddItem(ctx, userID, domain.Item{
			Sku:   sku,
			Count: 1,
		})
		require.NoError(t, err)
	}

	var wg sync.WaitGroup

	for _, sku := range skus {
		wg.Add(1)
		go func(s domain.Sku) {
			defer wg.Done()
			err := repo.DeleteItem(ctx, userID, s)
			require.NoError(t, err)
		}(sku)
	}

	wg.Wait()

	_, exists := repo.cartByUserID[userID]
	assert.False(t, exists, "repository.TestDeleteItem_Concurrent: cart should be deleted after all items removed")
}
