package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"uni-test-repo/pkg/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type WorkerConfig struct {
	PollInterval time.Duration
	BatchSize    int
}

// Fetch, publish and mark run in one transaction, so the claimed rows stay locked
// across the publish and anything not committed is retried on the next tick.
type Worker struct {
	transactor postgres.Transactor
	store      func(postgres.Executor) Store
	publisher  Publisher
	cfg        WorkerConfig
}

func NewWorker(transactor postgres.Transactor, store func(postgres.Executor) Store, publisher Publisher, cfg WorkerConfig) *Worker {
	return &Worker{
		transactor: transactor,
		store:      store,
		publisher:  publisher,
		cfg:        cfg,
	}
}

func (w *Worker) Start(ctx context.Context) {
	slog.Info("Outbox worker started",
		"poll_interval", w.cfg.PollInterval,
		"batch_size", w.cfg.BatchSize)

	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Outbox worker stopped")
			return
		case <-ticker.C:
			if err := w.poll(ctx); err != nil {
				slog.Error("Outbox poll failed", slog.Any("error", err))
			}
		}
	}
}

func (w *Worker) poll(ctx context.Context) error {
	return w.transactor.InTransaction(ctx, pgx.ReadCommitted, func(tx postgres.Executor) error {
		store := w.store(tx)

		events, err := store.FetchUnpublished(ctx, w.cfg.BatchSize)
		if err != nil {
			return fmt.Errorf("fetch unpublished: %w", err)
		}
		if len(events) == 0 {
			return nil
		}

		published := make([]uuid.UUID, 0, len(events))
		for _, e := range events {
			if err := w.publisher.Publish(ctx, e); err != nil {
				// Skip marking it; committed successes aren't re-sent, this one retries.
				slog.Warn("Publish outbox event failed",
					"id", e.ID, "event_type", e.EventType, slog.Any("error", err))
				continue
			}
			published = append(published, e.ID)
		}

		if err := store.MarkPublished(ctx, published); err != nil {
			return fmt.Errorf("mark published: %w", err)
		}
		return nil
	})
}
