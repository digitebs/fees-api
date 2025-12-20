package bill

import (
	"context"
	"errors"
	"time"

	"fees-api/money"
	"fees-api/temporal"
	"fees-api/types"

	"encore.dev/beta/errs"
	"encore.dev/config"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

type Config struct {
	TemporalServer string
}

var cfg = config.Load[*Config]()

func init() {
	temporal.InitClient(cfg.TemporalServer)
}

//encore:service
type Service struct{}

func NewBill(ctx context.Context, currency money.Currency) (*Bill, error) {
	total, err := money.NewMoney(0, currency)
	if err != nil {
		return nil, err
	}
	bill := &Bill{
		ID:        uuid.NewString(),
		Total:     total,
		Status:    Open,
		CreatedAt: time.Now(),
	}

	if err := CreateBill(ctx, bill); err != nil {
		return nil, err
	}

	var temporalClient = temporal.GetClient()

	// Start Temporal workflow
	_, err = temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        "bill-" + bill.ID,
			TaskQueue: "BILLING_TASK_QUEUE",
		},
		"BillWorkflow",
		bill.ID,
		currency,
	)

	return bill, nil
}

func EnsureOpen(b *Bill) error {
	if b.Status == Closed {
		return errors.New("bill already closed")
	}
	return nil
}

func CloseBill(ctx context.Context, id string) error {
	bill, err := GetBill(ctx, id)
	if err != nil {
		return errs.Wrap(err, "error fetching bill")
	}

	if err := EnsureOpen(bill); err != nil {
		return errs.Wrap(err, "bill is not open")
	}

	temporalClient := temporal.GetClient()

	err = temporalClient.SignalWorkflow(
		ctx,
		"bill-"+bill.ID,
		"",
		"close-bill",
		nil,
	)
	if err != nil {
		return errs.Wrap(err, "failed to signal close bill workflow")
	}

	return nil
}

func AddItemToBill(ctx context.Context, id, description string, amount int64) error {
	b, err := GetBill(ctx, id)
	if err != nil {
		return errs.Wrap(err, "bill not found")
	}

	if err := EnsureOpen(b); err != nil {
		return errs.Wrap(err, "failed to add item workflow")
	}

	temporalClient := temporal.GetClient()

	signal := types.AddItemSignal{
		ItemID:      uuid.NewString(),
		Amount:      amount,
		Description: description,
	}

	err = temporalClient.SignalWorkflow(
		ctx,
		"bill-"+b.ID,
		"",
		"add-item",
		signal,
	)
	if err != nil {
		return errs.Wrap(err, "failed to add item workflow")
	}

	return nil
}
