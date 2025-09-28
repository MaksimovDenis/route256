package order

import (
	"route256/loms/internal/domain"
	"sync"
)

type orderByID = map[int64]domain.Order

type sequenceGenerator interface {
	Add(delta int64) (updated int64)
}

type Repository struct {
	orderByID orderByID
	mx        sync.RWMutex

	sequenceGenerator sequenceGenerator
}

func New(sequenceGenerator sequenceGenerator) *Repository {
	return &Repository{
		orderByID:         make(orderByID),
		sequenceGenerator: sequenceGenerator,
	}
}

func (r *Repository) CreateOrder(order domain.Order) (int64, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	id := r.sequenceGenerator.Add(1)

	r.orderByID[id] = order

	return id, nil
}

func (r *Repository) GetByOrderID(orderID int64) (domain.Order, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	order, ok := r.orderByID[orderID]
	if !ok {
		return domain.Order{}, domain.ErrOrderNotFound
	}

	return order, nil
}

func (r *Repository) SetStatus(orderID int64, status domain.OrderStatus) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	order := r.orderByID[orderID]

	order.Status = status
	r.orderByID[orderID] = order

	return nil
}
