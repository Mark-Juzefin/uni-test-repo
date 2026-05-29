package product

import (
	"context"

	"github.com/google/uuid"
)

// ProductRepo is the persistence contract for products.
type ProductRepo interface {
	Create(ctx context.Context, p Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]Product, int, error)
}
