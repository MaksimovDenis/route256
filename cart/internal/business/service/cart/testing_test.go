package cart_test

import (
	cartservice "route256/cart/internal/business/service/cart"
	"route256/cart/internal/business/service/cart/mock"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
)

type fixture struct {
	*assert.Assertions

	productClient *mock.ProductClientMock
	lomsClient    *mock.LomsClientMock
	cartRepo      *mock.RepositoryMock

	executor *cartservice.Service
}

func setUp(t *testing.T) *fixture {
	ctrl := minimock.NewController(t)

	productClient := mock.NewProductClientMock(ctrl)
	lomsClient := mock.NewLomsClientMock(ctrl)
	cartRepo := mock.NewRepositoryMock(ctrl)

	executor := cartservice.New(
		cartRepo,
		productClient,
		lomsClient,
		5,
	)

	return &fixture{
		Assertions: assert.New(t),

		productClient: productClient,
		lomsClient:    lomsClient,
		cartRepo:      cartRepo,

		executor: executor,
	}
}
