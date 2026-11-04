package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/sohan-14/practice/internal/billing/domain"
)

// Common application errors.
var (
	ErrInvoiceNotFound = errors.New("invoice not found")
	ErrInvalidInput    = errors.New("invalid input")
)

// InvoiceService handles use cases for the billing domain.
type InvoiceService struct {
	repo   domain.InvoiceRepository
	logger *slog.Logger
	nextID int // simple ID generator for in-memory impl
}

// NewInvoiceService creates a new InvoiceService.
func NewInvoiceService(repo domain.InvoiceRepository, logger *slog.Logger) *InvoiceService {
	return &InvoiceService{
		repo:   repo,
		logger: logger,
		nextID: 1,
	}
}

// CreateInvoice creates a new invoice with the given line items.
func (s *InvoiceService) CreateInvoice(ctx context.Context, input CreateInvoiceInput) (*InvoiceOutput, error) {
	// Generate ID
	id := fmt.Sprintf("inv_%d", s.nextID)
	s.nextID++

	// Parse currency
	currency := domain.Currency(input.Currency)

	// Create the aggregate
	invoice, err := domain.NewInvoice(id, input.CustomerID, currency)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}

	// Add line items (delegates validation to the domain)
	for _, li := range input.LineItems {
		unitPrice, err := domain.NewMoneyFromFloat(li.UnitPrice, currency)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
		}
		item, err := domain.NewLineItem(li.Description, li.Quantity, unitPrice)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
		}
		if err := invoice.AddLineItem(item); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
		}
	}

	// Persist
	if err := s.repo.Save(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to save invoice: %w", err)
	}

	return s.toOutput(invoice), nil
}

// GetInvoice retrieves an invoice by ID.
func (s *InvoiceService) GetInvoice(ctx context.Context, id string) (*InvoiceOutput, error) {
	invoice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}
	return s.toOutput(invoice), nil
}

// ListInvoices retrieves all invoices.
func (s *InvoiceService) ListInvoices(ctx context.Context, limit, offset int) ([]*InvoiceOutput, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	invoices, err := s.repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}

	var outputs []*InvoiceOutput
	for _, inv := range invoices {
		outputs = append(outputs, s.toOutput(inv))
	}
	return outputs, nil
}

// IssueInvoice transitions an invoice from draft to issued.
func (s *InvoiceService) IssueInvoice(ctx context.Context, id string) (*InvoiceOutput, error) {
	invoice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}

	if err := invoice.Issue(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}

	if err := s.repo.Save(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to save invoice: %w", err)
	}

	// Process domain events
	for _, event := range invoice.PullEvents() {
		s.handleDomainEvent(event)
	}

	return s.toOutput(invoice), nil
}

// PayInvoice transitions an invoice from issued to paid.
func (s *InvoiceService) PayInvoice(ctx context.Context, id string) (*InvoiceOutput, error) {
	invoice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}

	if err := invoice.Pay(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}

	if err := s.repo.Save(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to save invoice: %w", err)
	}

	// Process domain events
	for _, event := range invoice.PullEvents() {
		s.handleDomainEvent(event)
	}

	return s.toOutput(invoice), nil
}

// handleDomainEvent processes a domain event.
// In a real system, this would publish to an event bus.
func (s *InvoiceService) handleDomainEvent(event interface{}) {
	switch e := event.(type) {
	case domain.InvoiceIssuedEvent:
		s.logger.Info("invoice issued",
			"invoice_id", e.InvoiceID,
			"customer_id", e.CustomerID,
			"total", e.Total.String(),
		)
	case domain.InvoicePaidEvent:
		s.logger.Info("invoice paid",
			"invoice_id", e.InvoiceID,
			"customer_id", e.CustomerID,
			"total", e.Total.String(),
		)
	}
}

// toOutput converts a domain invoice to an output DTO.
func (s *InvoiceService) toOutput(inv *domain.Invoice) *InvoiceOutput {
	var lineItems []LineItemOutput
	for _, li := range inv.LineItems() {
		lineItems = append(lineItems, LineItemOutput{
			Description: li.Description(),
			Quantity:    li.Quantity(),
			UnitPrice:   li.UnitPrice().ToFloat(),
			Total:       li.Total().ToFloat(),
		})
	}

	output := &InvoiceOutput{
		ID:         inv.ID(),
		CustomerID: inv.CustomerID(),
		Status:     string(inv.Status()),
		Currency:   string(inv.Currency()),
		LineItems:  lineItems,
		Total:      inv.Total().ToFloat(),
		CreatedAt:  inv.CreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  inv.UpdatedAt().Format("2006-01-02T15:04:05Z"),
	}

	if inv.IssuedAt() != nil {
		s := inv.IssuedAt().Format("2006-01-02T15:04:05Z")
		output.IssuedAt = &s
	}
	if inv.PaidAt() != nil {
		s := inv.PaidAt().Format("2006-01-02T15:04:05Z")
		output.PaidAt = &s
	}

	return output
}
