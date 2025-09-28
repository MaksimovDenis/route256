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

func TestAddItem(t *testing.T) {
	t.Parallel()

	err := metrics.Init(context.Background())
	require.NoError(t, err)

	err = logger.Init(zapcore.DebugLevel)
	require.NoError(t, err)

	const (
		userID      = uint64(1)
		sku         = domain.Sku(100)
		numRoutines = 10
		countPerAdd = 2
	)

	repo := New(10)
	ctx := context.Background()
	item := domain.Item{
		Sku:   sku,
		Count: countPerAdd,
	}

	var wg sync.WaitGroup

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.AddItem(ctx, userID, item)
			require.NoError(t, err)
		}()
	}

	wg.Wait()

	assert.Len(t, repo.cartByUserID, 1, "repository.TestAddItem: map should contain one user")
	assert.Len(t, repo.cartByUserID[userID], 1, "repository.TestAddItem: user cart should contain one item")
	assert.Equal(t, uint32(numRoutines*countPerAdd), repo.cartByUserID[userID][sku].Count, "repository.TestAddItem: item count should match expected total")
}
