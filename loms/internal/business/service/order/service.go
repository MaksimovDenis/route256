package order

import (
	"context"
	"route256/loms/internal/domain"
	txmanager "route256/loms/internal/infra/tx_manager"
)

//go:generate rm -rf mock
//go:generate mkdir -p mock
//go:generate minimock -i * -o ./mock -s "_mock.go" -g
type orderRepository interface {
	CreateOrder(ctx context.Context, userID int64) (int64, error)
	CreateOrderItems(ctx context.Context, orderID int64, items []domain.Item) error
	GetByOrderID(ctx context.Context, orderID int64) (domain.Order, error)
	GetByOrderIDForUpdate(ctx context.Context, orderID int64) (domain.Order, error)
	SetStatus(ctx context.Context, orderID int64, status domain.OrderStatus) error
	SetStatusAndCreateEvent(ctx context.Context, orderID int64, status domain.OrderStatus, event domain.Event) error
}

type txManager interface {
	ReadCommitted(ctx context.Context, f txmanager.Handler) error
}

type stockService interface {
	Reserve(ctx context.Context, items []domain.Item) error
	ReserveRemove(ctx context.Context, items []domain.Item) error
	ReserveCancel(ctx context.Context, items []domain.Item) error
}

type eventRepository interface {
	CreateEvent(ctx context.Context, events domain.Event) error
}
type Service struct {
	orderTopic      string
	orderRepository orderRepository
	stockService    stockService
	txManagerMaster txManager
	eventRepository eventRepository
}

func New(
	orderTopic string,
	orderRepository orderRepository,
	eventRepository eventRepository,
	stockService stockService,
	txManagerMaster txManager,
) *Service {
	return &Service{
		orderTopic:      orderTopic,
		eventRepository: eventRepository,
		orderRepository: orderRepository,
		stockService:    stockService,
		txManagerMaster: txManagerMaster,
	}
}
