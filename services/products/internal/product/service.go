package product

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ProductService struct {
	repo ProductRepo
}

func NewProductService(repo ProductRepo) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Create(ctx context.Context, req CreateProductRequest) (Product, error) {
	now := time.Now().UTC()
	p := Product{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return Product{}, fmt.Errorf("create product: %w", err)
	}
	return p, nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

func (s *ProductService) List(ctx context.Context, req ListProductsRequest) (ListProductsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	items, total, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return ListProductsResponse{}, fmt.Errorf("list products: %w", err)
	}

	return ListProductsResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
