package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AggregateType string

const AggregateProduct AggregateType = "product"

type EventType string

const (
	EventProductCreated EventType = "product.created"
	EventProductDeleted EventType = "product.deleted"
)

type Event struct {
	ID            uuid.UUID
	AggregateType AggregateType
	AggregateID   uuid.UUID
	EventType     EventType
	Payload       json.RawMessage
	CreatedAt     time.Time
	PublishedAt   *time.Time
}

type Store interface {
	Create(ctx context.Context, e Event) error
	FetchUnpublished(ctx context.Context, limit int) ([]Event, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID) error
}

type Publisher interface {
	Publish(ctx context.Context, e Event) error
}
