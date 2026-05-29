package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

var _ Publisher = (*KafkaPublisher)(nil)

func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: topic,
			// Hash balancer + AggregateID key keeps one aggregate's events on one
			// partition, so they stay ordered.
			Balancer:               &kafka.Hash{},
			RequiredAcks:           kafka.RequireOne,
			BatchTimeout:           10 * time.Millisecond,
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *KafkaPublisher) Publish(ctx context.Context, e Event) error {
	value, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(e.AggregateID.String()),
		Value: value,
	})
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
