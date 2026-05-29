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
	ID            uuid.UUID       `json:"id"`
	AggregateType AggregateType   `json:"aggregate_type"`
	AggregateID   uuid.UUID       `json:"aggregate_id"`
	EventType     EventType       `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"created_at"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
}

type Store interface {
	Create(ctx context.Context, e Event) error
	FetchUnpublished(ctx context.Context, limit int) ([]Event, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID) error
}

type Publisher interface {
	Publish(ctx context.Context, e Event) error
}
