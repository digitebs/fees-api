package bill

import (
	"time"

	"fees-api/money"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type BillState struct {
	BillID string
	Total  money.Money
	Closed bool
}

// BillWorkflow manages the lifecycle of a bill, handling item additions and bill closure.
// It uses Temporal workflow patterns to ensure consistency and reliability.
func BillWorkflow(ctx workflow.Context, billID string, currency money.Currency) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting bill workflow", "billID", billID, "currency", currency)

	total, err := money.NewMoney(0, currency)
	if err != nil {
		logger.Error("Failed to create initial money", "error", err)
		return err
	}

	state := BillState{
		BillID: billID,
		Total:  total,
	}

	// Add retry policy for activities
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute,
		MaximumAttempts:    5,
	}
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	addItemCh := workflow.GetSignalChannel(ctx, "add-item")
	closeCh := workflow.GetSignalChannel(ctx, "close-bill")

	for {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(addItemCh, func(c workflow.ReceiveChannel, more bool) {
			var s AddItemSignal
			c.Receive(ctx, &s)

			// Add input validation
			if s.ItemID == "" || s.Amount <= 0 || s.Description == "" {
				workflow.GetLogger(ctx).Warn("invalid add item signal", "signal", s)
				return
			}

			if state.Closed {
				return
			}

			itemMoney, err := money.NewMoney(s.Amount, state.Total.Currency)
			if err != nil {
				workflow.GetLogger(ctx).Error("failed to create item money", "err", err)
				return
			}

			err = workflow.ExecuteActivity(
				ctx,
				PersistLineItemActivity,
				LineItemInput{
					ID:          s.ItemID,
					BillID:      state.BillID,
					Amount:      itemMoney,
					Description: s.Description,
				},
			).Get(ctx, nil)
			if err != nil {
				workflow.GetLogger(ctx).Error("failed to persist line item after retries", "err", err)
				// Do not update total to maintain consistency
				return
			}

			// Only update total after successful persistence
			newTotal, err := state.Total.Add(itemMoney)
			if err != nil {
				workflow.GetLogger(ctx).Error("currency mismatch", "err", err)
				return
			}

			state.Total = newTotal
		})

		selector.AddReceive(closeCh, func(c workflow.ReceiveChannel, more bool) {
			state.Closed = true
		})

		selector.Select(ctx)

		if state.Closed {
			break
		}
	}

	return workflow.ExecuteActivity(
		ctx,
		FinalizeBillActivity,
		state,
		time.Now(),
	).Get(ctx, nil)
}
