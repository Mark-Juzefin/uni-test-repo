package productrepo

import (
	"context"
	"fmt"

	"uni-test-repo/pkg/postgres"
	"uni-test-repo/services/products/internal/product"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PgProductRepo struct {
	db      postgres.Executor
	builder squirrel.StatementBuilderType
}

func New(pg *postgres.Postgres) product.ProductRepo {
	return newRepo(pg.Pool, pg.Builder)
}

func TxRepoFactory(builder squirrel.StatementBuilderType) func(postgres.Executor) product.ProductRepo {
	return func(exec postgres.Executor) product.ProductRepo {
		return newRepo(exec, builder)
	}
}

func newRepo(db postgres.Executor, builder squirrel.StatementBuilderType) *PgProductRepo {
	return &PgProductRepo{db: db, builder: builder}
}

func (r *PgProductRepo) Create(ctx context.Context, p product.Product) error {
	query, args, err := r.builder.Insert("products").
		Columns("id", "name", "description", "price", "created_at", "updated_at").
		Values(p.ID, p.Name, p.Description, p.Price, p.CreatedAt, p.UpdatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("insert product: %w", err)
	}
	return nil
}

func (r *PgProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := r.builder.Delete("products").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete query: %w", err)
	}

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return product.ErrNotFound
	}
	return nil
}

func (r *PgProductRepo) List(ctx context.Context, limit, offset int) ([]product.Product, int, error) {
	query, args, err := r.builder.
		Select("id", "name", "description", "price", "created_at", "updated_at").
		From("products").
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build list query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[product.Product])
	if err != nil {
		return nil, 0, fmt.Errorf("scan products: %w", err)
	}

	countQuery, _, err := r.builder.Select("COUNT(*)").From("products").ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build count query: %w", err)
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	return items, total, nil
}
