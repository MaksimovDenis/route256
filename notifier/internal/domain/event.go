package domain

import "time"

type OrderEvent struct {
	OrderID int64
	Status  string
	Moment  time.Time
}
