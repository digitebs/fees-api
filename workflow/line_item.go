package workflow

import (
	"context"
	"time"

	"fees-api/bill"
	"fees-api/money"
)

type LineItemInput struct {
	ID          string
	BillID      string
	Amount      money.Money
	Description string
}

func PersistLineItemActivity(ctx context.Context, in LineItemInput) error {
	item := &bill.LineItem{
		ID:          in.ID,
		BillID:      in.BillID,
		Amount:      in.Amount,
		Description: in.Description,
		CreatedAt:   time.Now(),
	}

	return bill.InsertLineItem(ctx, item)
}
