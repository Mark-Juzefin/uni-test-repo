package outbox

import (
	"context"
	"fmt"
	"time"

	"uni-test-repo/pkg/postgres"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type PgOutboxStore struct {
	db      postgres.Executor
	builder squirrel.StatementBuilderType
}

var _ Store = (*PgOutboxStore)(nil)

func NewPgOutboxStore(db postgres.Executor, builder squirrel.StatementBuilderType) *PgOutboxStore {
	return &PgOutboxStore{db: db, builder: builder}
}

func TxStoreFactory(builder squirrel.StatementBuilderType) func(postgres.Executor) Store {
	return func(exec postgres.Executor) Store {
		return NewPgOutboxStore(exec, builder)
	}
}

func (s *PgOutboxStore) Create(ctx context.Context, e Event) error {
	query, args, err := s.builder.Insert("outbox").
		Columns("id", "aggregate_type", "aggregate_id", "event_type", "payload", "created_at").
		Values(e.ID, e.AggregateType, e.AggregateID, e.EventType, []byte(e.Payload), e.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build outbox insert: %w", err)
	}

	if _, err := s.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}

func (s *PgOutboxStore) FetchUnpublished(ctx context.Context, limit int) ([]Event, error) {
	query, args, err := s.builder.
		Select("id", "aggregate_type", "aggregate_id", "event_type", "payload", "created_at").
		From("outbox").
		Where(squirrel.Eq{"published_at": nil}).
		OrderBy("created_at").
		Limit(uint64(limit)).
		Suffix("FOR UPDATE SKIP LOCKED").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build unpublished query: %w", err)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query unpublished events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.AggregateType, &e.AggregateID, &e.EventType, &e.Payload, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan outbox event: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox rows: %w", err)
	}

	return events, nil
}

func (s *PgOutboxStore) MarkPublished(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	query, args, err := s.builder.Update("outbox").
		Set("published_at", time.Now().UTC()).
		Where(squirrel.Eq{"id": ids}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build mark published: %w", err)
	}

	if _, err := s.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("mark events published: %w", err)
	}
	return nil
}
