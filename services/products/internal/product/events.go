package product

import (
	"encoding/json"
	"fmt"
	"time"

	"uni-test-repo/services/products/internal/outbox"

	"github.com/google/uuid"
)

// productDeletedPayload is the body of a product.deleted event — just the id.
type productDeletedPayload struct {
	ID uuid.UUID `json:"id"`
}

// newOutboxEvent builds an outbox row for a product change. The payload is
// whatever the caller marshals (the full product on create, the id on delete).
func newOutboxEvent(eventType outbox.EventType, aggregateID uuid.UUID, payload any) (outbox.Event, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return outbox.Event{}, fmt.Errorf("marshal %s payload: %w", eventType, err)
	}

	return outbox.Event{
		ID:            uuid.New(),
		AggregateType: outbox.AggregateProduct,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Payload:       data,
		CreatedAt:     time.Now().UTC(),
	}, nil
}
