package domain

import "context"

// InvoiceRepository is the port that the store layer implements.
// It is defined in the domain layer because it represents the domain's
// persistence needs, not its technical implementation.
type InvoiceRepository interface {
	Save(ctx context.Context, invoice *Invoice) error
	FindByID(ctx context.Context, id string) (*Invoice, error)
	FindByCustomer(ctx context.Context, customerID string) ([]*Invoice, error)
	FindAll(ctx context.Context, limit, offset int) ([]*Invoice, error)
}
