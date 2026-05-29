package notifier

import (
	"context"
	"encoding/json"
	"log/slog"

	"uni-test-repo/services/notifications/internal/event"
)

type Notifier struct{}

func New() *Notifier {
	return &Notifier{}
}

func (n *Notifier) Handle(ctx context.Context, key, value []byte) error {
	var e event.ProductEvent
	if err := json.Unmarshal(value, &e); err != nil {
		slog.Error("Skipping malformed event", "key", string(key), slog.Any("error", err))
		return nil
	}

	slog.Info("Product event received",
		"event_type", e.EventType,
		"aggregate_id", e.AggregateID,
		"payload", string(e.Payload),
	)
	return nil
}
