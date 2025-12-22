package bill

import (
	"context"
	"errors"
	"strings"

	"fees-api/money"

	"encore.dev/beta/errs"
	"github.com/google/uuid"
)

// validateUUID validates that a string is a valid UUID format
func validateUUID(id string) error {
	if id == "" {
		return errs.WrapCode(errors.New("id is required"), errs.InvalidArgument, "id is required")
	}
	if _, err := uuid.Parse(id); err != nil {
		return errs.WrapCode(errors.New("id must be a valid UUID"), errs.InvalidArgument, "id must be a valid UUID")
	}
	return nil
}

type CreateBillRequest struct {
	Currency money.Currency `json:"currency"`
}

type CreateBillResponse struct {
	Bill *Bill `json:"bill"`
}

type GetBillResponse struct {
	Bill      *Bill       `json:"bill"`
	LineItems []*LineItem `json:"line_items"`
}

// CreateBillAPI creates a new bill with the specified currency and starts a Temporal workflow.
// encore:api public method=POST path=/bills
func CreateBillAPI(
	ctx context.Context,
	req CreateBillRequest,
) (*CreateBillResponse, error) {
	// Validate currency before processing
	if !req.Currency.IsValid() {
		return nil, errs.WrapCode(errors.New("invalid currency"), errs.InvalidArgument, "invalid currency")
	}

	b, err := Create(ctx, req.Currency)
	if err != nil {
		return nil, err
	}

	return &CreateBillResponse{Bill: b}, nil
}

// GetBillAPI retrieves a bill and its line items by ID.
// encore:api public method=GET path=/bills/:id
func GetBillAPI(ctx context.Context, id string) (*GetBillResponse, error) {
	// Validate bill ID format
	if err := validateUUID(id); err != nil {
		return nil, err
	}

	b, err := GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := GetLineItems(ctx, id)
	if err != nil {
		return nil, err
	}

	return &GetBillResponse{
		b, items,
	}, nil
}

type ListBillsRequest struct {
	Status string `query:"status"`
}

type ListBillsResponse struct {
	Bills []*Bill `json:"bills"`
}

// encore:api public method=GET path=/bills
func ListBills(
	ctx context.Context,
	req ListBillsRequest,
) (*ListBillsResponse, error) {
	// Validate status parameter
	if req.Status != "" {
		if req.Status != "OPEN" && req.Status != "CLOSED" {
			return nil, errs.WrapCode(errors.New("status must be OPEN or CLOSED"), errs.InvalidArgument, "status must be OPEN or CLOSED")
		}
	}

	var bills []*Bill
	var err error

	if req.Status == "" {
		bills, err = ListBillsAll(ctx)
	} else {
		bills, err = ListBillsByStatus(ctx, Status(req.Status))
	}

	if err != nil {
		return &ListBillsResponse{}, err
	}

	return &ListBillsResponse{Bills: bills}, nil
}

//encore:api public method=POST path=/bills/:id/close
func CloseBillAPI(ctx context.Context, id string) error {
	// Validate bill ID format
	if err := validateUUID(id); err != nil {
		return err
	}

	return Close(ctx, id)
}

// MaxAmountCents represents the maximum allowed amount for a line item ($1M in cents)
const MaxAmountCents = 1_000_000_00

type AddItemRequest struct {
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

//encore:api public method=POST path=/bills/:id/items
func AddItem(ctx context.Context, id string, req AddItemRequest) error {
	// Validate bill ID format
	if err := validateUUID(id); err != nil {
		return err
	}

	// Sanitize and validate inputs
	req.Description = strings.TrimSpace(req.Description)

	if req.Amount <= 0 || req.Amount > MaxAmountCents {
		return errs.WrapCode(errors.New("amount must be positive and reasonable"), errs.InvalidArgument, "amount must be positive and reasonable")
	}
	if len(req.Description) == 0 || len(req.Description) > 500 {
		return errs.WrapCode(errors.New("description required and max 500 chars"), errs.InvalidArgument, "description required and max 500 chars")
	}

	return AddLineItem(ctx, id, req.Amount, req.Description)
}
