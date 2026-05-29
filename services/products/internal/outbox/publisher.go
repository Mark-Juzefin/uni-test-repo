package outbox

import (
	"context"
	"log/slog"
)

type LogPublisher struct{}

var _ Publisher = LogPublisher{}

func (LogPublisher) Publish(ctx context.Context, e Event) error {
	slog.Info("outbox event published",
		"id", e.ID,
		"event_type", e.EventType,
		"aggregate_id", e.AggregateID,
		"payload", string(e.Payload),
	)
	return nil
}
