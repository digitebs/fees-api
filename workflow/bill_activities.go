package workflow

import (
	"context"
	"time"

	"fees-api/bill"
)

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
