package domain

import "time"

type OrderEvent struct {
	OrderID int64
	Status  string
	Moment  time.Time
}

type Event struct {
	ID      int64
	Topic   string
	Key     string
	Payload []byte
	Status  EventStatus
}

type EventStatus string

const (
	EventStatusNew   EventStatus = "new"
	EventStatusSent  EventStatus = "sent"
	EventStatusError EventStatus = "error"
)
