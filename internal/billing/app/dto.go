package app

// CreateInvoiceInput is the input for creating an invoice.
type CreateInvoiceInput struct {
	CustomerID string `json:"customer_id"`
	Currency   string `json:"currency"`
	LineItems  []LineItemInput `json:"line_items"`
}

// LineItemInput represents a line item in create/update requests.
type LineItemInput struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// InvoiceOutput is the output representation of an invoice.
type InvoiceOutput struct {
	ID         string             `json:"id"`
	CustomerID string             `json:"customer_id"`
	Status     string             `json:"status"`
	Currency   string             `json:"currency"`
	LineItems  []LineItemOutput   `json:"line_items"`
	Total      float64            `json:"total"`
	IssuedAt   *string            `json:"issued_at,omitempty"`
	PaidAt     *string            `json:"paid_at,omitempty"`
	CreatedAt  string             `json:"created_at"`
	UpdatedAt  string             `json:"updated_at"`
}

// LineItemOutput represents a line item in responses.
type LineItemOutput struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}
