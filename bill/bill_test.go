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
			err := ensureOpen(tt.bill)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureOpen() error = %v, wantErr %v", err, tt.wantErr)
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

func TestStatusString(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   string
	}{
		{"open status", Open, "OPEN"},
		{"closed status", Closed, "CLOSED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("Status.String() = %v, want %v", string(tt.status), tt.want)
			}
		})
	}
}

func TestBillValidation(t *testing.T) {
	tests := []struct {
		name  string
		bill  *Bill
		valid bool
	}{
		{"valid bill", &Bill{ID: "test-id", Status: Open, Total: money.Money{Amount: 100, Currency: money.USD}}, true},
		{"nil bill", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ensureOpen function
			if tt.valid && tt.bill != nil {
				err := ensureOpen(tt.bill)
				if tt.name == "valid bill" && err != nil {
					t.Errorf("ensureOpen() should not error for valid open bill, got %v", err)
				}
			}
		})
	}
}
