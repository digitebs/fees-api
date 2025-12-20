package api

import (
	"context"
	"fees-api/bill"
)

//encore:api public method=POST path=/bills/:id/close
func Close(ctx context.Context, id string) error {
	return bill.Close(ctx, id)
}
