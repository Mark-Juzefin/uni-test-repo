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
