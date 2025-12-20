package api

import (
	"context"

	"fees-api/bill"
)

type ListBillsRequest struct {
	Status string `query:"status"`
}

type ListBillsResponse struct {
	Bills []*bill.Bill `json:"bills"`
}

// encore:api public method=GET path=/bills
func ListBills(
	ctx context.Context,
	req ListBillsRequest,
) (ListBillsResponse, error) {
	var bills []*bill.Bill
	var err error

	if req.Status == "" {
		bills, err = bill.ListBillsAll(ctx)
	} else {
		bills, err = bill.ListBillsByStatus(ctx, bill.Status(req.Status))
	}

	if err != nil {
		return ListBillsResponse{}, err
	}

	return ListBillsResponse{Bills: bills}, nil
}
