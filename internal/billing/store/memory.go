package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/sohan-14/practice/internal/billing/domain"
)

// MemoryRepository is an in-memory implementation of InvoiceRepository.
type MemoryRepository struct {
	mu       sync.RWMutex
	invoices map[string]*domain.Invoice
}

// NewMemoryRepository creates a new in-memory repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		invoices: make(map[string]*domain.Invoice),
	}
}

func (r *MemoryRepository) Save(ctx context.Context, invoice *domain.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invoices[invoice.ID()] = invoice
	return nil
}

func (r *MemoryRepository) FindByID(ctx context.Context, id string) (*domain.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	inv, exists := r.invoices[id]
	if !exists {
		return nil, fmt.Errorf("invoice not found: %s", id)
	}
	return inv, nil
}

func (r *MemoryRepository) FindByCustomer(ctx context.Context, customerID string) ([]*domain.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Invoice
	for _, inv := range r.invoices {
		if inv.CustomerID() == customerID {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (r *MemoryRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Invoice
	count := 0

	for _, inv := range r.invoices {
		if count >= offset && len(result) < limit {
			result = append(result, inv)
		}
		count++
	}
	return result, nil
}
