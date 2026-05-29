package outbox

import (
	"context"
	"fmt"

	"uni-test-repo/pkg/postgres"

	"github.com/Masterminds/squirrel"
)

type PgOutboxStore struct {
	db      postgres.Executor
	builder squirrel.StatementBuilderType
}

var _ Store = (*PgOutboxStore)(nil)

func NewPgOutboxStore(db postgres.Executor, builder squirrel.StatementBuilderType) *PgOutboxStore {
	return &PgOutboxStore{db: db, builder: builder}
}

// TxStoreFactory binds an outbox store to a caller-supplied Executor (typically a
// live transaction), so the event commits atomically with the data that produced it.
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
