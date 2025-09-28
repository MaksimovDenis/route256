package syncproducer

import (
	"context"
	"fmt"
	"route256/loms/internal/domain"
	"route256/loms/internal/infra/config"
	"route256/loms/internal/infra/logger"
	"route256/loms/internal/infra/metrics"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/opentracing/opentracing-go"
)

type Producer struct {
	prc        sarama.SyncProducer
	orderTopic string
}

func New(ctx context.Context, kafkaConfig config.Config) (
	*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = kafkaConfig.Kafka.RetryCountMsg
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	borkerList := strings.Split(kafkaConfig.Kafka.Brokers, ",")

	prc, err := sarama.NewSyncProducer(borkerList, config)
	if err != nil {
		return nil, fmt.Errorf("sarama.NewSyncProducer %v", err)
	}

	logger.Infof(ctx, "sync producer successfully created")

	producer := Producer{
		prc:        prc,
		orderTopic: kafkaConfig.Kafka.OrderTopic,
	}

	return &producer, nil
}

func (p *Producer) SendOrderEventsBatch(ctx context.Context, orderEvents []domain.Event) (successIDs []int64, errorIDs []int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "kafkaProducer.SendOrderEventsBatch")
	defer span.Finish()

	messages := make([]*sarama.ProducerMessage, 0, len(orderEvents))
	msgMap := make(map[*sarama.ProducerMessage]int64, len(orderEvents))

	for _, event := range orderEvents {
		msg := &sarama.ProducerMessage{
			Key:   sarama.StringEncoder(event.Key),
			Topic: p.orderTopic,
			Value: sarama.StringEncoder(event.Payload),
		}

		messages = append(messages, msg)
		msgMap[msg] = event.ID
	}

	start := time.Now()
	err = p.prc.SendMessages(messages)
	duration := time.Since(start).Seconds()

	if err == nil {
		metrics.IncKafkaProduceCounter(p.orderTopic, "ok")
		metrics.KafkaProduceDurationHistogram(p.orderTopic, "ok", duration)

		successIDs = make([]int64, 0, len(orderEvents))
		for _, event := range orderEvents {
			successIDs = append(successIDs, event.ID)
		}
		return successIDs, nil, nil
	}

	metrics.IncKafkaProduceCounter(p.orderTopic, "error")
	metrics.KafkaProduceDurationHistogram(p.orderTopic, "error", duration)

	producerErrors, ok := err.(sarama.ProducerErrors)
	if !ok {
		return nil, nil, fmt.Errorf("prc.SendMessages: %w", err)
	}

	failed := make(map[int64]struct{}, len(producerErrors))
	errorIDs = make([]int64, 0, len(producerErrors))

	for _, pe := range producerErrors {
		if id, exists := msgMap[pe.Msg]; exists {
			logger.Warnf(ctx, "failed to produce msg with id = %v", id)

			errorIDs = append(errorIDs, id)
			failed[id] = struct{}{}
		}
	}

	successIDs = make([]int64, 0, len(orderEvents)-len(errorIDs))
	for _, event := range orderEvents {
		if _, isFailed := failed[event.ID]; !isFailed {
			successIDs = append(successIDs, event.ID)
		}
	}

	return successIDs, errorIDs, nil
}

func (p *Producer) Close(ctx context.Context) error {
	if err := p.prc.Close(); err != nil {
		return fmt.Errorf("producer.Close: %w", err)
	}

	logger.Infof(ctx, "Kafka producer closed successfully")

	return nil
}
