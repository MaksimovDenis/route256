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

func TestGetCountItems(t *testing.T) {
	t.Parallel()

	err := metrics.Init(context.Background())
	require.NoError(t, err)

	err = logger.Init(zapcore.DebugLevel)
	require.NoError(t, err)

	repo := New(10)
	ctx := context.Background()

	items := []struct {
		userID uint64
		item   domain.Item
	}{
		{userID: 1, item: domain.Item{Sku: 101, Count: 2}},
		{userID: 1, item: domain.Item{Sku: 102, Count: 3}},
		{userID: 2, item: domain.Item{Sku: 201, Count: 1}},
	}

	var totalExpected uint32
	for _, v := range items {
		require.NoError(t, repo.AddItem(ctx, v.userID, v.item))
		totalExpected += v.item.Count
	}

	count := repo.GetCountItems()
	assert.Equal(t, totalExpected, count)
}

func TestGetCountItems_Concurrent(t *testing.T) {
	t.Parallel()

	err := metrics.Init(context.Background())
	require.NoError(t, err)

	err = logger.Init(zapcore.DebugLevel)
	require.NoError(t, err)

	repo := New(10)
	ctx := context.Background()

	const userID = uint64(1)
	initialItem := domain.Item{Sku: 100, Count: 1}
	require.NoError(t, repo.AddItem(ctx, userID, initialItem))

	var wg sync.WaitGroup
	const writers = 5
	const iterations = 20

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				err := repo.AddItem(ctx, userID, domain.Item{Sku: 100, Count: 1})
				require.NoError(t, err)
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			count := repo.GetCountItems()
			assert.GreaterOrEqual(t, count, uint32(1))
			time.Sleep(10 * time.Millisecond)
		}
	}()

	wg.Wait()

	finalCount := repo.GetCountItems()
	assert.Equal(t, uint32(1+writers*iterations), finalCount)
}
