package money

import (
	"errors"
	"fmt"
	"math"
)

type Currency string

const (
	USD Currency = "USD"
	GEL Currency = "GEL"
)

type Money struct {
	Amount   int64 // smallest unit, e.g. cents or tetri
	Currency Currency
}

// NewMoney creates a new Money instance
func NewMoney(amount int64, currency Currency) (Money, error) {
	if amount < 0 {
		return Money{}, errors.New("amount cannot be negative")
	}
	if !currency.IsValid() {
		return Money{}, fmt.Errorf("invalid currency: %s", currency)
	}
	return Money{Amount: amount, Currency: currency}, nil
}

// Add adds two Money values of the same currency
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("currency mismatch: %s vs %s", m.Currency, other.Currency)
	}
	if other.Amount > 0 && m.Amount > math.MaxInt64-other.Amount {
		return Money{}, errors.New("addition would overflow")
	}
	if other.Amount < 0 && m.Amount < math.MinInt64-other.Amount {
		return Money{}, errors.New("addition would underflow")
	}
	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}, nil
}

// Current String() method is basic - consider formatting
func (m Money) String() string {
	// For USD: "12.34 USD" instead of "1234 USD"
	if m.Currency == USD {
		dollars := m.Amount / 100
		cents := m.Amount % 100
		return fmt.Sprintf("$%d.%02d %s", dollars, cents, m.Currency)
	}
	return fmt.Sprintf("%d %s", m.Amount, m.Currency)
}

// Add validation for supported currencies
func (c Currency) IsValid() bool {
	return c == USD || c == GEL
}
