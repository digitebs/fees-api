package bill

import (
	"context"
	"errors"
	"sync"
	"time"

	"fees-api/money"
	"fees-api/temporal"
	"fees-api/types"

	"encore.dev/beta/errs"
	"encore.dev/config"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

var (
	temporalCfg      = config.Load[*Config]()
	temporalClient   client.Client
	temporalClientMu sync.Mutex
)

//encore:service
type Service struct{}

// Create creates a new bill with the specified currency and starts a Temporal workflow
func Create(ctx context.Context, currency money.Currency) (*Bill, error) {
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

	// Start Temporal workflow
	_, err = GetTemporalClient().ExecuteWorkflow(
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

// GetByID retrieves a bill by ID
func GetByID(ctx context.Context, billID string) (*Bill, error) {
	return GetBill(ctx, billID)
}

// GetLineItems retrieves line items for a bill
func GetLineItems(ctx context.Context, billID string) ([]*LineItem, error) {
	return ListLineItems(ctx, billID)
}

// Close closes a bill by signaling the Temporal workflow
func Close(ctx context.Context, billID string) error {
	bill, err := GetByID(ctx, billID)
	if err != nil {
		return errs.Wrap(err, "error fetching bill")
	}

	if err := ensureOpen(bill); err != nil {
		return errs.Wrap(err, "bill is not open")
	}

	err = GetTemporalClient().SignalWorkflow(
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

// AddLineItem adds a line item to a bill by signaling the Temporal workflow
func AddLineItem(ctx context.Context, billID string, amount int64, description string) error {
	bill, err := GetByID(ctx, billID)
	if err != nil {
		return errs.Wrap(err, "bill not found")
	}

	if err := ensureOpen(bill); err != nil {
		return errs.Wrap(err, "failed to add item workflow")
	}

	signal := types.AddItemSignal{
		ItemID:      uuid.NewString(),
		Amount:      amount,
		Description: description,
	}

	err = GetTemporalClient().SignalWorkflow(
		ctx,
		"bill-"+bill.ID,
		"",
		"add-item",
		signal,
	)
	if err != nil {
		return errs.Wrap(err, "failed to add item workflow")
	}

	return nil
}

// GetTemporalClient returns the temporal client initialized for this service
func GetTemporalClient() client.Client {
	temporalClientMu.Lock()
	defer temporalClientMu.Unlock()

	if temporalClient == nil {
		temporalClient = temporal.GetClient(temporalCfg.TemporalServer)
	}

	return temporalClient
}

// ensureOpen checks if a bill is open and returns an error if not
func ensureOpen(b *Bill) error {
	if b.Status == Closed {
		return errors.New("bill already closed")
	}
	return nil
}
