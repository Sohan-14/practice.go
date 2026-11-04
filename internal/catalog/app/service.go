package app

import (
	"context"
	"errors"
	"fmt"
)

// Common errors
var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidInput    = errors.New("invalid input")
)

// ProductService handles the business logic for products.
type ProductService struct {
	repo Repository
}

// Repository is the interface that ProductService depends on.
// This is the "port" that the store layer implements.
type Repository interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context, limit, offset int) ([]*Product, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id string) error
}

// NewProductService creates a new ProductService.
func NewProductService(repo Repository) *ProductService {
	return &ProductService{repo: repo}
}

// CreateProduct creates a new product.
func (s *ProductService) CreateProduct(ctx context.Context, input CreateProductInput) (*Product, error) {
	// Validate input
	if input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if input.Price < 0 {
		return nil, fmt.Errorf("%w: price must be non-negative", ErrInvalidInput)
	}

	product := &Product{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// GetProduct retrieves a product by ID.
func (s *ProductService) GetProduct(ctx context.Context, id string) (*Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrProductNotFound, id)
	}
	return product, nil
}

// ListProducts retrieves a list of products.
func (s *ProductService) ListProducts(ctx context.Context, limit, offset int) ([]*Product, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

// UpdateProduct updates an existing product.
func (s *ProductService) UpdateProduct(ctx context.Context, id string, input UpdateProductInput) (*Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrProductNotFound, id)
	}

	// Apply updates
	if input.Name != nil {
		if *input.Name == "" {
			return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidInput)
		}
		product.Name = *input.Name
	}
	if input.Description != nil {
		product.Description = *input.Description
	}
	if input.Price != nil {
		if *input.Price < 0 {
			return nil, fmt.Errorf("%w: price must be non-negative", ErrInvalidInput)
		}
		product.Price = *input.Price
	}

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

// DeleteProduct deletes a product.
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%w: %s", ErrProductNotFound, id)
	}
	return nil
}
