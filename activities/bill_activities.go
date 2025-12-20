package activities

import (
	"context"
	"time"

	"fees-api/bill"
	"fees-api/money"
)

type BillState struct {
	BillID string
	Total  money.Money
	Closed bool
}

func FinalizeBillActivity(ctx context.Context, state BillState, closedAt time.Time) error {
	b, err := bill.GetBill(ctx, state.BillID)
	if err != nil {
		return err
	}

	b.Status = bill.Closed
	b.Total = state.Total
	b.ClosedAt = &closedAt

	return bill.UpdateBill(ctx, b)
}
