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
	ID        string
	Status    Status
	Total     money.Money
	CreatedAt time.Time
	ClosedAt  *time.Time
}

type LineItem struct {
	ID          string
	BillID      string
	Amount      money.Money
	Description string
	CreatedAt   time.Time
}
