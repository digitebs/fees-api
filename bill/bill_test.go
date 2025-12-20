package bill

import (
	"testing"
	"time"

	"fees-api/money"
)

func TestEnsureOpen(t *testing.T) {
	openBill := &Bill{Status: Open}
	closedBill := &Bill{Status: Closed}

	tests := []struct {
		name    string
		bill    *Bill
		wantErr bool
	}{
		{"open bill", openBill, false},
		{"closed bill", closedBill, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsureOpen(tt.bill)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureOpen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBill_StatusString(t *testing.T) {
	b := &Bill{
		ID:        "test-id",
		Status:    Open,
		Total:     money.Money{Amount: 100, Currency: money.USD},
		CreatedAt: time.Now(),
	}

	if b.Status != Open {
		t.Errorf("Expected status Open, got %v", b.Status)
	}
}
