package consumer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

const commitTimeout = 5 * time.Second

type MessageHandler func(ctx context.Context, key, value []byte) error

type Consumer struct {
	reader *kafka.Reader
}

func New(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			Topic:       topic,
			GroupID:     groupID,
			MinBytes:    1,
			MaxBytes:    10 * 1024 * 1024,
			StartOffset: kafka.FirstOffset,
		}),
	}
}

// Start consumes until ctx is cancelled.
func (c *Consumer) Start(ctx context.Context, handler MessageHandler) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("fetch message: %w", err)
		}

		if err := handler(ctx, msg.Key, msg.Value); err != nil {
			slog.Error("Handler failed, not committing", "offset", msg.Offset, slog.Any("error", err))
			continue
		}

		// Commit on a fresh context so a processed message survives shutdown of ctx.
		commitCtx, cancel := context.WithTimeout(context.Background(), commitTimeout)
		err = c.reader.CommitMessages(commitCtx, msg)
		cancel()
		if err != nil {
			slog.Error("Commit failed", "offset", msg.Offset, slog.Any("error", err))
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
