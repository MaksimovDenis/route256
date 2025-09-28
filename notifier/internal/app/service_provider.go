package app

import (
	"context"
	kafkaConsumer "route256/notifier/internal/adapter/kafka/consumer"
	serviceEvent "route256/notifier/internal/business/service/event"
	"route256/notifier/internal/infra/closer"
	"route256/notifier/internal/infra/config"
	"route256/notifier/internal/infra/logger"
	"strings"
)

type serviceProvider struct {
	config config.Config

	orderСonsumer             *kafkaConsumer.Consumer
	orderConsumerGroupHandler *kafkaConsumer.GroupHandler

	serviceEvent *serviceEvent.Service
}

func newServiceProvider() *serviceProvider {
	return &serviceProvider{}
}

func (srv *serviceProvider) ServiceEvent(_ context.Context) *serviceEvent.Service {
	if srv.serviceEvent == nil {
		event := serviceEvent.New()

		srv.serviceEvent = event
	}

	return srv.serviceEvent
}

func (srv *serviceProvider) OrderConsumerGroupHandler(ctx context.Context) *kafkaConsumer.GroupHandler {
	if srv.orderConsumerGroupHandler == nil {
		srv.orderConsumerGroupHandler = kafkaConsumer.NewGroupHandler(
			srv.ServiceEvent(ctx),
		)
	}

	return srv.orderConsumerGroupHandler
}

func (srv *serviceProvider) OrderConsumer(ctx context.Context) *kafkaConsumer.Consumer {
	brokers := strings.Split(srv.config.Kafka.Brokers, ",")

	if srv.orderСonsumer == nil {
		consumer, err := kafkaConsumer.New(
			ctx,
			brokers,
			srv.config.Kafka.OrderConsumerGroupID,
			srv.config.Kafka.OrderTopic,
			srv.OrderConsumerGroupHandler(ctx),
		)
		if err != nil {
			logger.Fatalf(ctx, "consumer.New: failed to run consumer %v", err)
		}

		srv.orderСonsumer = consumer
		closer.Add(srv.orderСonsumer.Close)
	}

	return srv.orderСonsumer
}
