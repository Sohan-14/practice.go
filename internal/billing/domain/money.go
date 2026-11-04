package domain

import (
	"errors"
	"fmt"
	"math"
)

// Currency represents a supported currency.
type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	GBP Currency = "GBP"
)

// ValidCurrencies is the set of supported currencies.
var ValidCurrencies = map[Currency]bool{
	USD: true,
	EUR: true,
	GBP: true,
}

// Money is a value object representing an amount in a specific currency.
// It is immutable: all operations return a new Money instance.
type Money struct {
	amount   int64    // stored in cents to avoid float precision issues
	currency Currency
}

// NewMoney creates a new Money value object.
// Amount is in cents (e.g., $10.50 = 1050).
func NewMoney(amount int64, currency Currency) (Money, error) {
	if !ValidCurrencies[currency] {
		return Money{}, fmt.Errorf("invalid currency: %s", currency)
	}
	return Money{amount: amount, currency: currency}, nil
}

// NewMoneyFromFloat creates Money from a float amount (e.g., 10.50 for $10.50).
func NewMoneyFromFloat(amount float64, currency Currency) (Money, error) {
	if !ValidCurrencies[currency] {
		return Money{}, fmt.Errorf("invalid currency: %s", currency)
	}
	cents := int64(math.Round(amount * 100))
	return Money{amount: cents, currency: currency}, nil
}

// Amount returns the amount in cents.
func (m Money) Amount() int64 {
	return m.amount
}

// Currency returns the currency.
func (m Money) Currency() Currency {
	return m.currency
}

// ToFloat returns the amount as a float (e.g., 10.50).
func (m Money) ToFloat() float64 {
	return float64(m.amount) / 100.0
}

// Add returns a new Money that is the sum of m and other.
// Currencies must match.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.New("cannot add different currencies")
	}
	return Money{
		amount:   m.amount + other.amount,
		currency: m.currency,
	}, nil
}

// Multiply returns a new Money multiplied by the given quantity.
func (m Money) Multiply(quantity int) Money {
	return Money{
		amount:   m.amount * int64(quantity),
		currency: m.currency,
	}
}

// IsZero returns true if the amount is zero.
func (m Money) IsZero() bool {
	return m.amount == 0
}

// String returns a human-readable representation.
func (m Money) String() string {
	return fmt.Sprintf("%s %.2f", m.currency, m.ToFloat())
}
