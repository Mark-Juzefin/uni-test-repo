package productrepo

import (
	"context"
	"fmt"

	"uni-test-repo/pkg/postgres"
	"uni-test-repo/services/products/internal/product"
)

type PgProductRepo struct {
	pg *postgres.Postgres
}

func New(pg *postgres.Postgres) product.ProductRepo {
	return &PgProductRepo{pg: pg}
}

func (r *PgProductRepo) Create(ctx context.Context, p product.Product) error {
	query, args, err := r.pg.Builder.Insert("products").
		Columns("id", "name", "description", "price", "created_at", "updated_at").
		Values(p.ID, p.Name, p.Description, p.Price, p.CreatedAt, p.UpdatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	if _, err := r.pg.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("insert product: %w", err)
	}
	return nil
}
