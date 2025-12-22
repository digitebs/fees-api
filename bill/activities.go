package bill

import (
	"context"
	"time"

	"fees-api/money"
)

func FinalizeBillActivity(ctx context.Context, billID string, closedAt time.Time) error {
	// Only update status to CLOSED and closed_at (preserve existing total)
	return UpdateBillStatusOnly(ctx, billID, Closed, &closedAt)
}

type AddLineItemInput struct {
	ItemID      string
	BillID      string
	Amount      money.Money
	Description string
	CreatedAt   time.Time
}

func AddLineItemActivity(ctx context.Context, input AddLineItemInput) error {
	return InsertLineItemAndUpdateTotal(ctx, input.BillID, &LineItem{
		ID:          input.ItemID,
		BillID:      input.BillID,
		Amount:      input.Amount,
		Description: input.Description,
		CreatedAt:   time.Now(),
	})
}
