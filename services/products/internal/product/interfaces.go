package product

import "context"

// ProductRepo is the persistence contract for products.
type ProductRepo interface {
	Create(ctx context.Context, p Product) error
}
