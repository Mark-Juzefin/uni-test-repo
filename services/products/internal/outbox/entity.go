package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AggregateType identifies the kind of domain object an event is about.
type AggregateType string

const AggregateProduct AggregateType = "product"

// EventType is the concrete event name.
type EventType string

const (
	EventProductCreated EventType = "product.created"
	EventProductDeleted EventType = "product.deleted"
)

// Event is one row of the transactional outbox, written in the same transaction
// as the change that produced it.
type Event struct {
	ID            uuid.UUID
	AggregateType AggregateType
	AggregateID   uuid.UUID
	EventType     EventType
	Payload       json.RawMessage
	CreatedAt     time.Time
	PublishedAt   *time.Time
}

// Store persists outbox events.
type Store interface {
	Create(ctx context.Context, e Event) error
}
