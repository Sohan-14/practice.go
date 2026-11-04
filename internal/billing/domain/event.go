package domain

import "time"

// InvoiceIssuedEvent is emitted when an invoice is issued.
type InvoiceIssuedEvent struct {
	InvoiceID  string
	CustomerID string
	Total      Money
	IssuedAt   time.Time
}

// InvoicePaidEvent is emitted when an invoice is paid.
type InvoicePaidEvent struct {
	InvoiceID  string
	CustomerID string
	Total      Money
	PaidAt     time.Time
}

// DomainEvent is a marker interface for all domain events.
type DomainEvent interface {
	domainEvent()
}

func (InvoiceIssuedEvent) domainEvent() {}
func (InvoicePaidEvent) domainEvent()   {}
