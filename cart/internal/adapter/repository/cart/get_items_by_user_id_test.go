package cart

import (
	"context"
	"route256/cart/internal/domain"
	"route256/cart/internal/infra/logger"
	"route256/cart/internal/infra/metrics"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestGetItemsByUserID(t *testing.T) {
	t.Parallel()

	err := metrics.Init(context.Background())
	require.NoError(t, err)

	err = logger.Init(zapcore.DebugLevel)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      uint64
		addItem     *domain.Item
		want        []domain.Item
		wantLen     int
		expectedErr error
	}{
		{
			name:    "success: repository.GetItemsByUserID",
			userID:  123,
			addItem: &domain.Item{Sku: domain.Sku(456), Count: 2},
			want: []domain.Item{
				{
					Sku:   456,
					Count: 2,
				},
			},
			wantLen:     1,
			expectedErr: nil,
		},
		{
			name:        "fail: repository.GetItemsByUserID ErrEmptyCart",
			userID:      999,
			addItem:     nil,
			want:        nil,
			wantLen:     0,
			expectedErr: domain.ErrEmptyCart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := New(10)

			ctx := context.Background()

			if tt.addItem != nil {
				err := repo.AddItem(ctx, tt.userID, *tt.addItem)
				require.NoError(t, err)
			}

			items, err := repo.GetItemsByUserID(ctx, tt.userID)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			assert.Len(t, items, tt.wantLen)
			assert.Equal(t, items, tt.want)
		})
	}
}

func TestGetItemsByUserID_Concurrent(t *testing.T) {
	t.Parallel()

	err := metrics.Init(context.Background())
	require.NoError(t, err)

	err = logger.Init(zapcore.DebugLevel)
	require.NoError(t, err)

	const userID = uint64(1)

	repo := New(10)
	ctx := context.Background()

	initialItem := domain.Item{Sku: 100, Count: 1}
	err = repo.AddItem(ctx, userID, initialItem)
	require.NoError(t, err)

	var wg sync.WaitGroup

	readers := 5

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				items, err := repo.GetItemsByUserID(ctx, userID)
				require.NoError(t, err)

				assert.True(t, len(items) == 1)

				if len(items) == 1 {
					assert.Equal(t, initialItem.Sku, items[0].Sku)
					assert.GreaterOrEqual(t, items[0].Count, initialItem.Count)
				}

				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			err := repo.AddItem(ctx, userID, domain.Item{Sku: 100, Count: 1})
			require.NoError(t, err)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	wg.Wait()
}
