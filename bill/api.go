package bill

import (
	"context"
	"errors"

	"fees-api/money"
)

type CreateBillRequest struct {
	Currency money.Currency `json:"currency"`
}

type GetBillResponse struct {
	Bill      *Bill
	LineItems []*LineItem `json:"line_items"`
}

// CreateBillAPI creates a new bill with the specified currency and starts a Temporal workflow.
// encore:api public method=POST path=/bills
func CreateBillAPI(
	ctx context.Context,
	req CreateBillRequest,
) (*GetBillResponse, error) {

	b, err := Create(ctx, req.Currency)
	if err != nil {
		return &GetBillResponse{}, err
	}
	return &GetBillResponse{Bill: b}, nil
}

// GetBillAPI retrieves a bill and its line items by ID.
// encore:api public method=GET path=/bills/:id
func GetBillAPI(ctx context.Context, id string) (*GetBillResponse, error) {

	b, err := GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := GetLineItems(ctx, id)
	if err != nil {
		return nil, err
	}

	return &GetBillResponse{
		Bill:      b,
		LineItems: items,
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
	if req.Amount <= 0 || req.Amount > MaxAmountCents {
		return errors.New("amount must be positive and reasonable")
	}
	if len(req.Description) == 0 || len(req.Description) > 500 {
		return errors.New("description required and max 500 chars")
	}

	return AddLineItem(ctx, id, req.Amount, req.Description)
}
