package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func ConsumeKafkaMessages(broker, topic, groupID string, timeout time.Duration) ([]kafka.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         []string{broker},
		Topic:           topic,
		GroupID:         groupID,
		MinBytes:        10e3, // 10KB
		MaxBytes:        10e6, // 10MB
		StartOffset:     kafka.FirstOffset,
		ReadLagInterval: -1,
	})

	defer r.Close()

	var messages []kafka.Message

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			if err == context.DeadlineExceeded || err == context.Canceled {
				break
			}
			log.Printf("error while reading kafka message: %v", err)
			break
		}
		messages = append(messages, m)
		log.Printf("received message: %s\n", string(m.Value))
	}

	return messages, nil
}
