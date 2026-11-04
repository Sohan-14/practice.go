package domain

import (
	"errors"
	"fmt"
	"time"
)

// InvoiceStatus represents the state of an invoice.
type InvoiceStatus string

const (
	StatusDraft     InvoiceStatus = "draft"
	StatusIssued    InvoiceStatus = "issued"
	StatusPaid      InvoiceStatus = "paid"
	StatusCancelled InvoiceStatus = "cancelled"
)

// Common domain errors (these are part of the domain's contract).
var (
	ErrInvalidStatus    = errors.New("invalid status transition")
	ErrCannotModifyPaid = errors.New("cannot modify a paid invoice")
	ErrNoLineItems      = errors.New("invoice must have at least one line item")
	ErrCurrencyMismatch = errors.New("all line items must have the same currency")
)

// Invoice is the aggregate root for the billing domain.
// It enforces all business invariants.
type Invoice struct {
	id         string
	customerID string
	status     InvoiceStatus
	currency   Currency
	lineItems  []LineItem
	total      Money
	issuedAt   *time.Time
	paidAt     *time.Time
	createdAt  time.Time
	updatedAt  time.Time
	events     []interface{} // domain events
}

// NewInvoice creates a new invoice in draft status.
func NewInvoice(id, customerID string, currency Currency) (*Invoice, error) {
	if id == "" {
		return nil, fmt.Errorf("invoice ID is required")
	}
	if customerID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}
	if !ValidCurrencies[currency] {
		return nil, fmt.Errorf("invalid currency: %s", currency)
	}

	now := time.Now()
	zero, _ := NewMoney(0, currency)

	inv := &Invoice{
		id:         id,
		customerID: customerID,
		status:     StatusDraft,
		currency:   currency,
		lineItems:  make([]LineItem, 0),
		total:      zero,
		createdAt:  now,
		updatedAt:  now,
	}

	return inv, nil
}

// AddLineItem adds a line item to the invoice.
// Invariant: Can only add items to draft invoices.
// Invariant: All items must have the same currency as the invoice.
func (i *Invoice) AddLineItem(item LineItem) error {
	if i.status != StatusDraft {
		return fmt.Errorf("%w: can only add items to draft invoices", ErrCannotModifyPaid)
	}
	if item.UnitPrice().Currency() != i.currency {
		return ErrCurrencyMismatch
	}

	i.lineItems = append(i.lineItems, item)
	i.total, _ = i.total.Add(item.Total())
	i.updatedAt = time.Now()

	return nil
}

// RemoveLineItem removes a line item by index.
func (i *Invoice) RemoveLineItem(index int) error {
	if i.status != StatusDraft {
		return fmt.Errorf("%w: can only remove items from draft invoices", ErrCannotModifyPaid)
	}
	if index < 0 || index >= len(i.lineItems) {
		return fmt.Errorf("invalid line item index: %d", index)
	}

	i.lineItems = append(i.lineItems[:index], i.lineItems[index+1:]...)

	newTotal, _ := NewMoney(0, i.currency)
	for _, li := range i.lineItems {
		newTotal, _ = newTotal.Add(li.Total())
	}
	i.total = newTotal
	i.updatedAt = time.Now()

	return nil
}

// Issue transitions the invoice from draft to issued.
// Invariant: Must have at least one line item.
func (i *Invoice) Issue() error {
	if i.status != StatusDraft {
		return fmt.Errorf("%w: can only issue draft invoices", ErrInvalidStatus)
	}
	if len(i.lineItems) == 0 {
		return ErrNoLineItems
	}

	i.status = StatusIssued
	now := time.Now()
	i.issuedAt = &now
	i.updatedAt = now

	i.events = append(i.events, InvoiceIssuedEvent{
		InvoiceID:  i.id,
		CustomerID: i.customerID,
		Total:      i.total,
		IssuedAt:   now,
	})

	return nil
}

// Pay transitions the invoice from issued to paid.
// Invariant: Must be issued first.
func (i *Invoice) Pay() error {
	if i.status != StatusIssued {
		return fmt.Errorf("%w: can only pay issued invoices", ErrInvalidStatus)
	}

	i.status = StatusPaid
	now := time.Now()
	i.paidAt = &now
	i.updatedAt = now

	i.events = append(i.events, InvoicePaidEvent{
		InvoiceID:  i.id,
		CustomerID: i.customerID,
		Total:      i.total,
		PaidAt:     now,
	})

	return nil
}

// Cancel cancels the invoice.
// Invariant: Cannot cancel paid invoices.
func (i *Invoice) Cancel() error {
	if i.status == StatusPaid {
		return fmt.Errorf("%w: cannot cancel paid invoices", ErrInvalidStatus)
	}
	if i.status == StatusCancelled {
		return fmt.Errorf("%w: invoice is already cancelled", ErrInvalidStatus)
	}

	i.status = StatusCancelled
	i.updatedAt = time.Now()

	return nil
}

// PullEvents returns and clears pending domain events.
func (i *Invoice) PullEvents() []interface{} {
	events := i.events
	i.events = nil
	return events
}

// Getters (read-only access to internal state)

func (i *Invoice) ID() string            { return i.id }
func (i *Invoice) CustomerID() string    { return i.customerID }
func (i *Invoice) Status() InvoiceStatus { return i.status }
func (i *Invoice) Currency() Currency    { return i.currency }
func (i *Invoice) LineItems() []LineItem { return i.lineItems }
func (i *Invoice) Total() Money          { return i.total }
func (i *Invoice) IssuedAt() *time.Time  { return i.issuedAt }
func (i *Invoice) PaidAt() *time.Time    { return i.paidAt }
func (i *Invoice) CreatedAt() time.Time  { return i.createdAt }
func (i *Invoice) UpdatedAt() time.Time  { return i.updatedAt }
