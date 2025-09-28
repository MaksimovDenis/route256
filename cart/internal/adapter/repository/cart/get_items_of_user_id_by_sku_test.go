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

func TestGetItemOfUserIDBySku_Concurrent(t *testing.T) {
	t.Parallel()

	err := metrics.Init(context.Background())
	require.NoError(t, err)

	err = logger.Init(zapcore.DebugLevel)
	require.NoError(t, err)

	const userID = uint64(1)
	const sku = domain.Sku(100)

	repo := New(10)
	ctx := context.Background()

	initialItem := domain.Item{Sku: sku, Count: 1}
	err = repo.AddItem(ctx, userID, initialItem)
	require.NoError(t, err)

	var wg sync.WaitGroup

	readers := 5
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				item, err := repo.GetItemOfUserIDBySku(ctx, userID, sku)
				if err != nil {
					require.ErrorIs(t, err, domain.ErrItemNotFound)
				} else {
					assert.Equal(t, sku, item.Sku)
					assert.GreaterOrEqual(t, item.Count, uint32(1))
				}
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			err := repo.AddItem(ctx, userID, domain.Item{Sku: sku, Count: 1})
			require.NoError(t, err)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	wg.Wait()
}
