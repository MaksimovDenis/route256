package orderevent_test

import (
	orderevent "route256/loms/internal/business/cron/order_event"
	"route256/loms/internal/business/cron/order_event/mock"
	"route256/loms/internal/infra/logger"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

type fixture struct {
	*assert.Assertions
	eventRepository *mock.EventRepositoryMock
	producerKafka   *mock.ProducerKafkaMock
	txManager       *mock.TxManagerMock
	executor        *orderevent.CronProcessor
}

func setUp(t *testing.T) *fixture {
	ctrl := minimock.NewController(t)

	err := logger.Init(zapcore.DebugLevel)
	require.NoError(t, err)

	eventRepository := mock.NewEventRepositoryMock(ctrl)
	producerMock := mock.NewProducerKafkaMock(ctrl)
	txManagerMock := mock.NewTxManagerMock(ctrl)

	executor := orderevent.New(eventRepository, producerMock, txManagerMock, 100)

	return &fixture{
		Assertions:      assert.New(t),
		eventRepository: eventRepository,
		producerKafka:   producerMock,
		txManager:       txManagerMock,
		executor:        executor,
	}
}
