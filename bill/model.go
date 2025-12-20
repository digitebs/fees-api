package bill

import (
	"time"

	"fees-api/money"
)

type Status string

const (
	Open   Status = "OPEN"
	Closed Status = "CLOSED"
)

type Bill struct {
	ID        string      `json:"id"`
	Status    Status      `json:"status"`
	Total     money.Money `json:"total"`
	CreatedAt time.Time   `json:"created_at"`
	ClosedAt  *time.Time  `json:"closed_at,omitempty"` // omit if nil
}

type LineItem struct {
	ID          string      `json:"id"`
	BillID      string      `json:"bill_id"`
	Amount      money.Money `json:"amount"`
	Description string      `json:"description"`
	CreatedAt   time.Time   `json:"created_at"`
}
