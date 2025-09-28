package stock_test

import (
	stockservice "route256/loms/internal/business/service/stock"
	"route256/loms/internal/business/service/stock/mock"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
)

type fixture struct {
	*assert.Assertions

	repository *mock.StockRepositoryMock

	executor *stockservice.Service
}

func setUp(t *testing.T) *fixture {
	ctrl := minimock.NewController(t)

	repository := mock.NewStockRepositoryMock(ctrl)

	executor := stockservice.New(repository)

	return &fixture{
		Assertions: assert.New(t),

		repository: repository,

		executor: executor,
	}
}
