package product

import "errors"

var (
	ErrNotFound      = errors.New("product not found")
	ErrAlreadyExists = errors.New("product already exists")
	ErrInvalidInput  = errors.New("invalid product input")
)
