package api

import (
	"context"
	"errors"
	"fees-api/bill"
)

type AddItemRequest struct {
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

//encore:api public method=POST path=/bills/:id/items
func AddItem(ctx context.Context, id string, req AddItemRequest) error {
	if req.Amount <= 0 || req.Amount > 1_000_000_00 { // Max $1M in cents
		return errors.New("amount must be positive and reasonable")
	}
	if len(req.Description) == 0 || len(req.Description) > 500 {
		return errors.New("description required and max 500 chars")
	}

	return bill.AddItemToBill(ctx, id, req.Description, req.Amount)
}
