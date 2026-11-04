package domain

import "fmt"

// LineItem is a value object representing a single line on an invoice.
type LineItem struct {
	description string
	quantity    int
	unitPrice   Money
}

// NewLineItem creates a new line item.
func NewLineItem(description string, quantity int, unitPrice Money) (LineItem, error) {
	if description == "" {
		return LineItem{}, fmt.Errorf("description is required")
	}
	if quantity <= 0 {
		return LineItem{}, fmt.Errorf("quantity must be positive")
	}
	if unitPrice.Amount() < 0 {
		return LineItem{}, fmt.Errorf("unit price must be non-negative")
	}
	return LineItem{
		description: description,
		quantity:    quantity,
		unitPrice:   unitPrice,
	}, nil
}

// Description returns the description.
func (li LineItem) Description() string {
	return li.description
}

// Quantity returns the quantity.
func (li LineItem) Quantity() int {
	return li.quantity
}

// UnitPrice returns the unit price.
func (li LineItem) UnitPrice() Money {
	return li.unitPrice
}

// Total returns the total for this line (quantity × unit price).
func (li LineItem) Total() Money {
	return li.unitPrice.Multiply(li.quantity)
}
