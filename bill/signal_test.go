package bill

import (
	"testing"
)

func TestAddItemSignal(t *testing.T) {
	signal := AddItemSignal{
		ItemID:      "test-id",
		Amount:      100,
		Description: "test item",
	}

	if signal.ItemID != "test-id" {
		t.Errorf("Expected ItemID test-id, got %v", signal.ItemID)
	}
	if signal.Amount != 100 {
		t.Errorf("Expected Amount 100, got %v", signal.Amount)
	}
	if signal.Description != "test item" {
		t.Errorf("Expected Description test item, got %v", signal.Description)
	}
}
