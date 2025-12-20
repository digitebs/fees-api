package api

import (
	"context"

	"fees-api/bill"
	"fees-api/money"
)

type CreateBillRequest struct {
	Currency money.Currency `json:"currency"`
}

type GetBillResponse struct {
	Bill      *bill.Bill
	LineItems []*bill.LineItem `json:"line_items"`
}

// Create creates a new bill with the specified currency and starts a Temporal workflow.
// encore:api public method=POST path=/bills
func Create(
	ctx context.Context,
	req CreateBillRequest,
) (GetBillResponse, error) {

	b, err := bill.Create(ctx, req.Currency)
	if err != nil {
		return GetBillResponse{}, err
	}
	return GetBillResponse{Bill: b}, nil
}

// GetBillAPI retrieves a bill and its line items by ID.
// encore:api public method=GET path=/bills/:id
func GetBillAPI(ctx context.Context, id string) (*GetBillResponse, error) {

	b, err := bill.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := bill.GetLineItems(ctx, id)
	if err != nil {
		return nil, err
	}

	return &GetBillResponse{
		Bill:      b,
		LineItems: items,
	}, nil
}
