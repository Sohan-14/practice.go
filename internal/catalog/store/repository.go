package store

import (
	"context"

	"github.com/sohan-14/practice/internal/catalog/app"
)

// Repository defines the interface for product data access.
// This is the "port" that the app layer owns.
type Repository interface {
	Create(ctx context.Context, product *app.Product) error
	GetByID(ctx context.Context, id string) (*app.Product, error)
	List(ctx context.Context, limit, offset int) ([]*app.Product, error)
	Update(ctx context.Context, product *app.Product) error
	Delete(ctx context.Context, id string) error
}
