package money

import (
	"testing"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		currency Currency
		wantErr  bool
	}{
		{"valid USD", 100, USD, false},
		{"valid GEL", 50, GEL, false},
		{"negative amount", -10, USD, true},
		{"invalid currency", 100, Currency("INVALID"), true},
		{"zero amount", 0, USD, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMoney(tt.amount, tt.currency)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMoney() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMoney_Add(t *testing.T) {
	m1, _ := NewMoney(100, USD)
	m2, _ := NewMoney(50, USD)
	m3, _ := NewMoney(100, GEL)

	tests := []struct {
		name    string
		m       Money
		other   Money
		want    int64
		wantErr bool
	}{
		{"same currency", m1, m2, 150, false},
		{"different currency", m1, m3, 0, true},
		{"zero add", m1, Money{Amount: 0, Currency: USD}, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.Add(tt.other)
			if (err != nil) != tt.wantErr {
				t.Errorf("Money.Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Amount != tt.want {
				t.Errorf("Money.Add() = %v, want %v", got.Amount, tt.want)
			}
		})
	}
}

func TestMoney_String(t *testing.T) {
	tests := []struct {
		name string
		m    Money
		want string
	}{
		{"USD cents", Money{Amount: 123, Currency: USD}, "$1.23 USD"},
		{"USD dollars", Money{Amount: 200, Currency: USD}, "$2.00 USD"},
		{"GEL", Money{Amount: 100, Currency: GEL}, "100 GEL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.String(); got != tt.want {
				t.Errorf("Money.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrency_IsValid(t *testing.T) {
	tests := []struct {
		name string
		c    Currency
		want bool
	}{
		{"USD", USD, true},
		{"GEL", GEL, true},
		{"invalid", Currency("EUR"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.IsValid(); got != tt.want {
				t.Errorf("Currency.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
