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
	*bill.Bill
	LineItems []*bill.LineItem `json:"line_items"`
}

// encore:api public method=POST path=/bills
func Create(
	ctx context.Context,
	req CreateBillRequest,
) (GetBillResponse, error) {

	b, err := bill.NewBill(ctx, req.Currency)
	if err != nil {
		return GetBillResponse{}, err
	}
	if err != nil {
		return GetBillResponse{}, err
	}

	return GetBillResponse{Bill: b}, nil
}

// encore:api public method=GET path=/bills/:id
func GetBillAPI(ctx context.Context, id string) (*GetBillResponse, error) {

	b, err := bill.GetBill(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := bill.ListLineItems(ctx, id)
	if err != nil {
		return nil, err
	}

	return &GetBillResponse{
		Bill:      b,
		LineItems: items,
	}, nil
}
