package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sohan-14/practice/internal/catalog/app"
)

// MemoryRepository is an in-memory implementation of Repository.
type MemoryRepository struct {
	mu       sync.RWMutex
	products map[string]*app.Product
	nextID   int
}

// NewMemoryRepository creates a new in-memory repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		products: make(map[string]*app.Product),
		nextID:   1,
	}
}

func (r *MemoryRepository) Create(ctx context.Context, product *app.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	product.ID = fmt.Sprintf("prod_%d", r.nextID)
	r.nextID++
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	r.products[product.ID] = product
	return nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, id string) (*app.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	product, exists := r.products[id]
	if !exists {
		return nil, fmt.Errorf("product not found: %s", id)
	}
	return product, nil
}

func (r *MemoryRepository) List(ctx context.Context, limit, offset int) ([]*app.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*app.Product
	count := 0

	for _, product := range r.products {
		if count >= offset && len(result) < limit {
			result = append(result, product)
		}
		count++
	}

	return result, nil
}

func (r *MemoryRepository) Update(ctx context.Context, product *app.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.products[product.ID]; !exists {
		return fmt.Errorf("product not found: %s", product.ID)
	}

	product.UpdatedAt = time.Now()
	r.products[product.ID] = product
	return nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.products[id]; !exists {
		return fmt.Errorf("product not found: %s", id)
	}

	delete(r.products, id)
	return nil
}
