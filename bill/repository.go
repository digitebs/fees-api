package bill

import (
	"context"
	"database/sql"
	"errors"

	"fees-api/money"

	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("fees", sqldb.DatabaseConfig{
	Migrations: "./db/migrations",
})

func getDB() *sqldb.Database {
	return db
}

func CreateBill(ctx context.Context, bill *Bill) error {
	_, err := getDB().Exec(ctx, `
        INSERT INTO bills (id, currency, status, total_amount, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `, bill.ID, bill.Total.Currency, bill.Status, bill.Total.Amount, bill.CreatedAt)
	return err
}

var ErrBillNotFound = errors.New("bill not found")

func GetBill(ctx context.Context, billID string) (*Bill, error) {
	row := getDB().QueryRow(ctx, `
        SELECT
            id,
            currency,
            status,
            total_amount,
            created_at,
            closed_at
        FROM bills
        WHERE id = $1
    `, billID)

	var (
		b           Bill
		currencyStr string
		totalAmount int64
		closedAt    sql.NullTime
	)

	err := row.Scan(
		&b.ID,
		&currencyStr,
		&b.Status,
		&totalAmount,
		&b.CreatedAt,
		&closedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrBillNotFound
	}
	if err != nil {
		return nil, err
	}

	// reconstruct domain money
	m, err := money.NewMoney(totalAmount, money.Currency(currencyStr))
	if err != nil {
		return nil, err
	}
	b.Total = m

	if closedAt.Valid {
		b.ClosedAt = &closedAt.Time
	}

	return &b, nil
}

func UpdateBill(ctx context.Context, b *Bill) error {
	_, err := getDB().Exec(ctx, `
        UPDATE bills SET status=$1, total_amount=$2, closed_at=$3 WHERE id=$4
    `, b.Status, b.Total.Amount, b.ClosedAt, b.ID)
	return err
}

func ListBillsByStatus(
	ctx context.Context,
	status Status,
) ([]*Bill, error) {

	rows, err := getDB().Query(ctx, `
        SELECT
            id,
            currency,
            status,
            total_amount,
            created_at,
            closed_at
        FROM bills
        WHERE status = $1
        ORDER BY created_at DESC
    `, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bills []*Bill

	for rows.Next() {
		var (
			b           Bill
			currencyStr string
			totalAmount int64
			closedAt    sql.NullTime
		)

		if err := rows.Scan(
			&b.ID,
			&currencyStr,
			&b.Status,
			&totalAmount,
			&b.CreatedAt,
			&closedAt,
		); err != nil {
			return nil, err
		}

		m, err := money.NewMoney(totalAmount, money.Currency(currencyStr))
		if err != nil {
			return nil, err
		}
		b.Total = m

		if closedAt.Valid {
			b.ClosedAt = &closedAt.Time
		}

		bills = append(bills, &b)
	}

	return bills, nil
}

func ListBillsAll(ctx context.Context) ([]*Bill, error) {
	rows, err := getDB().Query(ctx, `
        SELECT
            id,
            currency,
            status,
            total_amount,
            created_at,
            closed_at
        FROM bills
        ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bills []*Bill

	for rows.Next() {
		var (
			b           Bill
			currencyStr string
			totalAmount int64
			closedAt    sql.NullTime
		)

		if err := rows.Scan(
			&b.ID,
			&currencyStr,
			&b.Status,
			&totalAmount,
			&b.CreatedAt,
			&closedAt,
		); err != nil {
			return nil, err
		}

		m, err := money.NewMoney(totalAmount, money.Currency(currencyStr))
		if err != nil {
			return nil, err
		}
		b.Total = m

		if closedAt.Valid {
			b.ClosedAt = &closedAt.Time
		}

		bills = append(bills, &b)
	}

	return bills, nil
}

func InsertLineItem(ctx context.Context, item *LineItem) error {
	_, err := getDB().Exec(ctx, `
        INSERT INTO line_items (
            id, bill_id, amount, currency, description, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6)
    `,
		item.ID,
		item.BillID,
		item.Amount.Amount,
		item.Amount.Currency,
		item.Description,
		item.CreatedAt,
	)
	return err
}

func ListLineItems(ctx context.Context, billID string) ([]*LineItem, error) {
	rows, err := getDB().Query(ctx, `
        SELECT id, amount, currency, description, created_at
        FROM line_items
        WHERE bill_id = $1
        ORDER BY created_at ASC
    `, billID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*LineItem

	for rows.Next() {
		var (
			li          LineItem
			amount      int64
			currencyStr string
		)

		if err := rows.Scan(
			&li.ID,
			&amount,
			&currencyStr,
			&li.Description,
			&li.CreatedAt,
		); err != nil {
			return nil, err
		}

		li.BillID = billID
		m, err := money.NewMoney(amount, money.Currency(currencyStr))
		if err != nil {
			return nil, err
		}
		li.Amount = m

		items = append(items, &li)
	}

	return items, nil
}
