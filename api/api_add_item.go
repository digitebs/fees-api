package api

import (
	"context"
	"errors"
	"fees-api/bill"
)

// MaxAmountCents represents the maximum allowed amount for a line item ($1M in cents)
const MaxAmountCents = 1_000_000_00

type AddItemRequest struct {
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

//encore:api public method=POST path=/bills/:id/items
func AddItem(ctx context.Context, id string, req AddItemRequest) error {
	if req.Amount <= 0 || req.Amount > MaxAmountCents {
		return errors.New("amount must be positive and reasonable")
	}
	if len(req.Description) == 0 || len(req.Description) > 500 {
		return errors.New("description required and max 500 chars")
	}

	return bill.AddLineItem(ctx, id, req.Amount, req.Description)
}
