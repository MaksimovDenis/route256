package cart_test

import (
	"context"
	cartrepository "route256/cart/internal/adapter/repository/cart"
	"route256/cart/internal/domain"
	"route256/cart/internal/infra/logger"
	"route256/cart/internal/infra/metrics"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func BenchmarkAddItem(b *testing.B) {
	err := metrics.Init(context.Background())
	require.NoError(b, err)

	err = logger.Init(zapcore.DebugLevel)
	require.NoError(b, err)

	repo := cartrepository.New(100)
	var userID uint64 = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item := domain.Item{
			Sku:   domain.Sku(i), // #nosec G115
			Count: 1,
		}

		_ = repo.AddItem(context.Background(), userID, item)
	}
}

func BenchmarkGetItemsByUserID(b *testing.B) {
	err := metrics.Init(context.Background())
	require.NoError(b, err)

	err = logger.Init(zapcore.DebugLevel)
	require.NoError(b, err)

	count := 100000
	repo := newFilledRepository(count)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetItemsByUserID(context.Background(), uint64(i)) // #nosec G115
	}
}

func BenchmarkDeleteItem(b *testing.B) {
	err := metrics.Init(context.Background())
	require.NoError(b, err)

	err = logger.Init(zapcore.DebugLevel)
	require.NoError(b, err)

	count := 100000
	repo := newFilledRepository(count)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.DeleteItem(context.Background(), uint64(i), domain.Sku(i)) // #nosec G115
	}
}

func BenchmarkDeleteItemsByUserID(b *testing.B) {
	err := metrics.Init(context.Background())
	require.NoError(b, err)

	err = logger.Init(zapcore.DebugLevel)
	require.NoError(b, err)

	var userID uint64 = 1

	count := 1000

	repo := newFilledRepository(count)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.DeleteItemsByUserID(context.Background(), userID)
	}

}

func newFilledRepository(count int) *cartrepository.Repository {
	repo := cartrepository.New(100)
	for i := 0; i < count; i++ {
		_ = repo.AddItem(context.Background(), uint64(i), domain.Item{ // #nosec G115
			Sku:   domain.Sku(i), // #nosec G115
			Count: 1,
		})
	}
	return repo
}
