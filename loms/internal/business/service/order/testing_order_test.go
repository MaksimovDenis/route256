package order_test

import (
	orderservice "route256/loms/internal/business/service/order"
	"route256/loms/internal/business/service/order/mock"

	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
)

type fixture struct {
	*assert.Assertions
	orderRepository *mock.OrderRepositoryMock
	eventRepository *mock.EventRepositoryMock

	stockService *mock.StockServiceMock

	txManager *mock.TxManagerMock

	executor *orderservice.Service
}

func setUp(t *testing.T) *fixture {
	ctrl := minimock.NewController(t)

	orderRepository := mock.NewOrderRepositoryMock(ctrl)
	eventRepository := mock.NewEventRepositoryMock(ctrl)
	txManager := mock.NewTxManagerMock(ctrl)
	stockService := mock.NewStockServiceMock(ctrl)

	executor := orderservice.New("test-topic", orderRepository, eventRepository, stockService, txManager)

	return &fixture{
		Assertions:      assert.New(t),
		orderRepository: orderRepository,
		eventRepository: eventRepository,
		stockService:    stockService,
		txManager:       txManager,
		executor:        executor,
	}
}
