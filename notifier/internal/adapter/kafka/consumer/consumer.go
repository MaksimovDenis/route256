package consumer

import (
	"context"
	"errors"
	"fmt"
	"route256/notifier/internal/infra/logger"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumerGroup        sarama.ConsumerGroup
	consumerGroupHandler *GroupHandler
	topicName            string
	errorHandler         sync.Once
}

func New(
	ctx context.Context,
	brokers []string,
	groupID string,
	topic string,
	consumerGroupHandler *GroupHandler,
) (*Consumer, error) {
	config := sarama.NewConfig()

	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = time.Second
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("sarama.NewConsumerGroup failed to create consumer group: %w", err)
	}

	logger.Infof(ctx, "Consumer group successfully created for groupID=%s topic=%s", groupID, topic)

	consumer := &Consumer{
		consumerGroup:        consumerGroup,
		consumerGroupHandler: consumerGroupHandler,
		topicName:            topic,
	}

	return consumer, nil
}

func (c *Consumer) startErrorHandler(ctx context.Context) {
	c.errorHandler.Do(func() {
		go func() {
			for err := range c.consumerGroup.Errors() {
				logger.Errorf(ctx, "Consumer group error: %v", err)
			}
		}()
	})
}

func (c *Consumer) Consume(ctx context.Context) error {
	c.startErrorHandler(ctx)
	return c.consume(ctx)
}

func (c *Consumer) Close() error {
	return c.consumerGroup.Close()
}

func (c *Consumer) consume(ctx context.Context) error {
	for {
		err := c.consumerGroup.Consume(ctx, []string{c.topicName}, c.consumerGroupHandler)
		if err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) {
				return nil
			}

			return fmt.Errorf("consumerGroup.Consume: failed to consume message: %w", err)
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		logger.Infof(ctx, "Rebalancing consumer group for topic %s", c.topicName)
	}
}
