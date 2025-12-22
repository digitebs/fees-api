package bill

import (
	"errors"
	"strings"
	"time"

	"fees-api/money"

	"github.com/google/uuid"
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

			// Enhanced signal validation
			if err := validateAddItemSignal(s); err != nil {
				workflow.GetLogger(ctx).Error("invalid add item signal", "error", err, "signal", s)
				return
			}

			if state.Closed {
				workflow.GetLogger(ctx).Warn("attempted to add item to closed bill", "billID", state.BillID)
				return
			}

			itemMoney, err := money.NewMoney(s.Amount, state.Total.Currency)
			if err != nil {
				workflow.GetLogger(ctx).Error("failed to create item money", "err", err)
				return
			}

			// Use transactional activity that handles both line item insertion and total update atomically
			err = workflow.ExecuteActivity(
				ctx,
				AddLineItemActivity,
				AddLineItemInput{
					ItemID:      s.ItemID,
					BillID:      state.BillID,
					Amount:      itemMoney,
					Description: s.Description,
					CreatedAt:   workflow.Now(ctx), // Use workflow time for determinism
				},
			).Get(ctx, nil)
			if err != nil {
				workflow.GetLogger(ctx).Error("failed to add line item transactionally", "err", err)
				return
			}

			// Update workflow state after successful transactional activity
			// Add the new item amount to the running total
			newTotal, err := state.Total.Add(itemMoney)
			if err != nil {
				workflow.GetLogger(ctx).Error("failed to add item to total", "err", err)
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
		state.BillID,
		workflow.Now(ctx), // Use workflow time for determinism
	).Get(ctx, nil)
}

// validateAddItemSignal performs comprehensive validation of add item signals
func validateAddItemSignal(s AddItemSignal) error {
	// Validate ItemID format (must be valid UUID)
	if s.ItemID == "" {
		return errors.New("item ID is required")
	}
	if _, err := uuid.Parse(s.ItemID); err != nil {
		return errors.New("item ID must be a valid UUID")
	}

	// Validate amount (must be positive and within business limits)
	if s.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	if s.Amount > 1_000_000_00 { // $1M limit
		return errors.New("amount exceeds maximum allowed ($1M)")
	}

	// Validate description (required and reasonable length)
	trimmedDesc := strings.TrimSpace(s.Description)
	if trimmedDesc == "" {
		return errors.New("description is required")
	}
	if len(trimmedDesc) > 500 {
		return errors.New("description exceeds maximum length (500 characters)")
	}

	// Additional business validations could be added here:
	// - Check for duplicate ItemIDs (would require state tracking)
	// - Validate amount precision (must be valid cents)
	// - Check business rules (hourly limits, etc.)

	return nil
}
