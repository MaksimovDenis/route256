package domain

type Sku int64

type Item struct {
	Sku   Sku
	Count int64
}

type Order struct {
	UserID int64
	Status OrderStatus
	Items  []Item
}

type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "new"
	OrderStatusAwaitingPayment OrderStatus = "awaiting payment"
	OrderStatusFailed          OrderStatus = "failed"
	OrderStatusPayed           OrderStatus = "payed"
	OrderStatusCancelled       OrderStatus = "cancelled"
)
