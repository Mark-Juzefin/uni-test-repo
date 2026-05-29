package postgres

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool    *pgxpool.Pool
	Builder squirrel.StatementBuilderType
}

func New(ctx context.Context, url string, maxPool int) (*Postgres, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse pg url: %w", err)
	}
	cfg.MaxConns = int32(maxPool)

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pg pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pg: %w", err)
	}

	return &Postgres{
		Pool:    pool,
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

// Executor is the common subset of *pgxpool.Pool and pgx.Tx, so a repository
// runs unchanged on the pool or inside a transaction.
type Executor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type Transactor interface {
	InTransaction(ctx context.Context, isoLevel pgx.TxIsoLevel, fn func(tx Executor) error) error
}

// The named return drives the deferred rollback: any error from fn or the commit
// rolls back, while a clean commit leaves err == nil and skips it.
func (p *Postgres) InTransaction(ctx context.Context, isoLevel pgx.TxIsoLevel, fn func(tx Executor) error) (err error) {
	tx, err := p.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: isoLevel})
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
