package cart

import (
	"context"
	"route256/cart/internal/domain"
)

//go:generate rm -rf mock
//go:generate mkdir -p mock
//go:generate minimock -i * -o ./mock -s "_mock.go" -g
type repository interface {
	GetItemsByUserID(ctx context.Context, userID uint64) ([]domain.Item, error)
	AddItem(ctx context.Context, userID uint64, cart domain.Item) error
	DeleteItem(ctx context.Context, userID uint64, sku domain.Sku) error
	DeleteItemsByUserID(ctx context.Context, userID uint64) error
	GetItemOfUserIDBySku(ctx context.Context, userID uint64, sku domain.Sku) (domain.Item, error)
}

type productClient interface {
	GetProductBySku(ctx context.Context, sku domain.Sku) (domain.Product, error)
}

type lomsClient interface {
	OrderCreate(ctx context.Context, userID uint64, items []domain.CartItem) (int64, error)
	StocksInfo(ctx context.Context, sku uint64) (int64, error)
}

type Service struct {
	repository    repository
	productClient productClient
	lomsClient    lomsClient
	workersCount  int
}

func New(
	repository repository,
	productClient productClient,
	lomsClient lomsClient,
	workersCount int,
) *Service {
	return &Service{
		repository:    repository,
		productClient: productClient,
		lomsClient:    lomsClient,
		workersCount:  workersCount,
	}
}
