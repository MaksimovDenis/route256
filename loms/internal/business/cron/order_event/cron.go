package orderevent

import (
	"context"
	"route256/loms/internal/domain"
	txmanager "route256/loms/internal/infra/tx_manager"
)

//go:generate rm -rf mock
//go:generate mkdir -p mock
//go:generate minimock -i * -o ./mock -s "_mock.go" -g
type producerKafka interface {
	SendOrderEventsBatch(ctx context.Context, orderEvents []domain.Event) (successIDs []int64, errorIDs []int64, err error)
}

type eventRepository interface {
	FetchNextMessages(ctx context.Context, limit int32) ([]domain.Event, error)
	MarkAsSent(ctx context.Context, ids []int64) error
	MarkAsError(ctx context.Context, orderIDs []int64) error
}

type txManager interface {
	ReadCommitted(ctx context.Context, f txmanager.Handler) error
}

type CronProcessor struct {
	eventRepository eventRepository
	producerKafka   producerKafka
	txManagerMaster txManager
	limitMsg        int32
}

func New(
	eventRepository eventRepository,
	producerKafka producerKafka,
	txManagerMaster txManager,
	limitMsg int32,
) *CronProcessor {
	return &CronProcessor{
		eventRepository: eventRepository,
		producerKafka:   producerKafka,
		txManagerMaster: txManagerMaster,
		limitMsg:        limitMsg,
	}
}
