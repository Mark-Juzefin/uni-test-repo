package product

import (
	"time"

	"github.com/google/uuid"
)

// Product is the core aggregate of the products domain.
type Product struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int64     `json:"price"` // minor units (e.g. cents)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProductRequest is the input DTO for product creation.
type CreateProductRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=2000"`
	Price       int64  `json:"price" binding:"required,gte=0"`
}
