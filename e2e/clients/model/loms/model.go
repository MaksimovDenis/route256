package loms

import "net/http"

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

type Stock struct {
	TotalCount int64
	Reserved   int64
}

type RespWithData[T any] struct {
	HTTPResp *http.Response
	Data     T
}
