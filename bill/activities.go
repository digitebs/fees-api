package bill

import (
	"context"
	"time"

	"fees-api/money"
)

func FinalizeBillActivity(ctx context.Context, state BillState, closedAt time.Time) error {
	b, err := GetBill(ctx, state.BillID)
	if err != nil {
		return err
	}

	b.Status = Closed
	b.Total = state.Total
	b.ClosedAt = &closedAt

	return UpdateBill(ctx, b)
}

type LineItemInput struct {
	ID          string
	BillID      string
	Amount      money.Money
	Description string
}

func PersistLineItemActivity(ctx context.Context, in LineItemInput) error {
	item := &LineItem{
		ID:          in.ID,
		BillID:      in.BillID,
		Amount:      in.Amount,
		Description: in.Description,
		CreatedAt:   time.Now(),
	}

	return InsertLineItem(ctx, item)
}
