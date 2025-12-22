package bill

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"fees-api/money"

	"encore.dev/beta/errs"
	"encore.dev/config"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

var (
	temporalCfg    = config.Load[*Config]()
	temporalClient client.Client
	temporalOnce   sync.Once
)

//encore:service
type Service struct{}

//encore:api private
func (s *Service) HealthCheck(ctx context.Context) error {
	// Could add actual health checks here
	return nil
}

// Create creates a new bill with the specified currency and starts a Temporal workflow
func Create(ctx context.Context, currency money.Currency) (*Bill, error) {
	// Check if Temporal is available
	if GetTemporalClient() == nil {
		return nil, errs.WrapCode(nil, errs.Unavailable,
			"bill creation unavailable - Temporal workflow service is down")
	}

	billID := uuid.NewString()

	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute,
		MaximumAttempts:    3,
	}

	// Start Temporal workflow FIRST to avoid race condition
	_, err := GetTemporalClient().ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:          "bill-" + billID,
			TaskQueue:   "BILLING_TASK_QUEUE",
			RetryPolicy: retryPolicy,
		},
		"BillWorkflow",
		billID,
		currency,
	)
	if err != nil {
		return nil, errs.Wrap(err, "failed to start bill workflow")
	}

	// Create bill in database AFTER workflow successfully started
	total, err := money.NewMoney(0, currency)
	if err != nil {
		return nil, errs.Wrap(err, "failed to create initial money amount")
	}

	bill := &Bill{
		ID:        billID,
		Total:     total,
		Status:    Open,
		CreatedAt: time.Now(),
	}
	if err := CreateBill(ctx, bill); err != nil {
		// Workflow started but bill creation failed
		_ = GetTemporalClient().TerminateWorkflow(ctx, "bill-"+billID, "", "Database creation failed", nil)
		return nil, errs.Wrap(err, "workflow started but bill creation failed")
	}

	return bill, nil
}

// GetByID retrieves a bill by ID
func GetByID(ctx context.Context, billID string) (*Bill, error) {
	bill, err := GetBill(ctx, billID)
	if err != nil {
		return nil, errs.WrapCode(err, errs.NotFound, "bill not found")
	}
	return bill, nil
}

// GetLineItems retrieves line items for a bill
func GetLineItems(ctx context.Context, billID string) ([]*LineItem, error) {
	return ListLineItems(ctx, billID)
}

// Close closes a bill by signaling the Temporal workflow
func Close(ctx context.Context, billID string) error {
	// Check if Temporal is available
	if GetTemporalClient() == nil {
		return errs.WrapCode(nil, errs.Unavailable,
			"bill operations unavailable - Temporal workflow service is down")
	}

	bill, err := GetByID(ctx, billID)
	if err != nil {
		return err
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
	// Check if Temporal is available
	if GetTemporalClient() == nil {
		return errs.WrapCode(nil, errs.Unavailable,
			"bill operations unavailable - Temporal workflow service is down")
	}

	bill, err := GetByID(ctx, billID)
	if err != nil {
		return err
	}

	if err := ensureOpen(bill); err != nil {
		return errs.Wrap(err, "failed to add item workflow")
	}

	signal := AddItemSignal{
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
// Returns nil if Temporal server is unavailable (logs warning)
func GetTemporalClient() client.Client {
	temporalOnce.Do(func() {
		client, err := client.Dial(client.Options{HostPort: temporalCfg.TemporalServer})
		if err != nil {
			log.Printf("WARNING: Temporal server unavailable at %s: %v. Bill workflow operations will fail.",
				temporalCfg.TemporalServer, err)
			return // temporalClient remains nil
		}
		temporalClient = client
	})
	return temporalClient
}

// ensureOpen checks if a bill is open and returns an error if not
func ensureOpen(b *Bill) error {
	if b.Status == Closed {
		return errors.New("bill already closed")
	}
	return nil
}
