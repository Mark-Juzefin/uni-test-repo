package product

import (
	"context"
	"fmt"
	"time"

	"uni-test-repo/pkg/postgres"
	"uni-test-repo/services/products/internal/outbox"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProductService struct {
	repo       ProductRepo                          // pool-bound, for non-tx reads (List)
	transactor postgres.Transactor                  // opens transactions
	txRepo     func(postgres.Executor) ProductRepo  // binds the product repo to a tx
	txOutbox   func(postgres.Executor) outbox.Store // binds the outbox store to a tx
}

func NewProductService(
	repo ProductRepo,
	transactor postgres.Transactor,
	txRepo func(postgres.Executor) ProductRepo,
	txOutbox func(postgres.Executor) outbox.Store,
) *ProductService {
	return &ProductService{
		repo:       repo,
		transactor: transactor,
		txRepo:     txRepo,
		txOutbox:   txOutbox,
	}
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

	evt, err := newOutboxEvent(outbox.EventProductCreated, p.ID, p)
	if err != nil {
		return Product{}, err
	}

	// Product and event commit together; no contended row, so ReadCommitted suffices.
	err = s.transactor.InTransaction(ctx, pgx.ReadCommitted, func(tx postgres.Executor) error {
		if err := s.txRepo(tx).Create(ctx, p); err != nil {
			return fmt.Errorf("create product: %w", err)
		}
		if err := s.txOutbox(tx).Create(ctx, evt); err != nil {
			return fmt.Errorf("write outbox event: %w", err)
		}
		return nil
	})
	if err != nil {
		return Product{}, err
	}
	return p, nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	evt, err := newOutboxEvent(outbox.EventProductDeleted, id, productDeletedPayload{ID: id})
	if err != nil {
		return err
	}

	return s.transactor.InTransaction(ctx, pgx.ReadCommitted, func(tx postgres.Executor) error {
		if err := s.txRepo(tx).Delete(ctx, id); err != nil {
			return fmt.Errorf("delete product: %w", err)
		}
		if err := s.txOutbox(tx).Create(ctx, evt); err != nil {
			return fmt.Errorf("write outbox event: %w", err)
		}
		return nil
	})
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
