package bill

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"fees-api/money"

	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("fees", sqldb.DatabaseConfig{
	Migrations: "./db/migrations",
})

// scanBill reconstructs a Bill domain object from database row data
func scanBill(row interface{ Scan(...interface{}) error }) (*Bill, error) {
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

func CreateBill(ctx context.Context, bill *Bill) error {
	_, err := db.Exec(ctx, `
        INSERT INTO bills (id, currency, status, total_amount, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `, bill.ID, bill.Total.Currency, bill.Status, bill.Total.Amount, bill.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create bill %s: %w", bill.ID, err)
	}
	return nil
}

var ErrBillNotFound = errors.New("bill not found")

func GetBill(ctx context.Context, billID string) (*Bill, error) {
	row := db.QueryRow(ctx, `
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

	bill, err := scanBill(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("bill not found for id %s: %w", billID, ErrBillNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan bill for id %s: %w", billID, err)
	}

	return bill, nil
}

func UpdateBill(ctx context.Context, b *Bill) error {
	_, err := db.Exec(ctx, `
        UPDATE bills SET status=$1, total_amount=$2, closed_at=$3 WHERE id=$4
    `, b.Status, b.Total.Amount, b.ClosedAt, b.ID)
	if err != nil {
		return fmt.Errorf("failed to update bill %s: %w", b.ID, err)
	}
	return nil
}

// UpdateBillTransactional updates bill status, total, and closed_at in a transaction
func UpdateBillTransactional(ctx context.Context, billID string, total money.Money, status Status, closedAt *time.Time) error {
	_, err := db.Exec(ctx, `
        UPDATE bills SET status=$1, total_amount=$2, closed_at=$3 WHERE id=$4
    `, status, total.Amount, closedAt, billID)
	if err != nil {
		return fmt.Errorf("failed to update bill %s transactionally: %w", billID, err)
	}
	return nil
}

// UpdateBillStatusOnly updates only bill status and closed_at (preserves existing total)
func UpdateBillStatusOnly(ctx context.Context, billID string, status Status, closedAt *time.Time) error {
	_, err := db.Exec(ctx, `
        UPDATE bills SET status=$1, closed_at=$2 WHERE id=$3
    `, status, closedAt, billID)
	if err != nil {
		return fmt.Errorf("failed to update bill %s status: %w", billID, err)
	}
	return nil
}

// InsertLineItemAndUpdateTotal inserts a line item and updates the bill total atomically in a single transaction.
// This function is idempotent - it can be called multiple times safely.
func InsertLineItemAndUpdateTotal(ctx context.Context, billID string, item *LineItem) error {
	return insertLineItemAndUpdateTotalTx(ctx, billID, item)
}

// insertLineItemAndUpdateTotalTx performs the atomic insert and total update within a transaction
func insertLineItemAndUpdateTotalTx(ctx context.Context, billID string, item *LineItem) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for bill %s: %w", billID, err)
	}
	defer tx.Rollback()

	// Check if line item already exists for idempotency (within transaction for atomicity)
	existing, err := getLineItemByIDTx(ctx, tx, billID, item.ID)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil // Idempotent - item already exists, no changes needed
	}

	// Get current bill total within transaction
	currentTotal, err := getBillTotalTx(ctx, tx, billID)
	if err != nil {
		return err
	}

	// Calculate new total
	newTotal, err := currentTotal.Add(item.Amount)
	if err != nil {
		return fmt.Errorf("failed to calculate new total for bill %s: %w", billID, err)
	}

	// Insert line item and update total atomically
	if err := InsertLineItemTx(ctx, tx, item); err != nil {
		return err
	}

	if err := updateBillTotalTx(ctx, tx, billID, newTotal.Amount); err != nil {
		return err
	}

	return tx.Commit()
}

// getBillTotalTx retrieves the current total for a bill within a transaction
func getBillTotalTx(ctx context.Context, tx *sqldb.Tx, billID string) (money.Money, error) {
	row := tx.QueryRow(ctx, `
		SELECT currency, total_amount FROM bills WHERE id = $1
	`, billID)

	var currencyStr string
	var totalAmount int64

	if err := row.Scan(&currencyStr, &totalAmount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return money.Money{}, fmt.Errorf("bill not found for id %s: %w", billID, ErrBillNotFound)
		}
		return money.Money{}, fmt.Errorf("failed to get bill total for id %s: %w", billID, err)
	}

	return money.NewMoney(totalAmount, money.Currency(currencyStr))
}

// updateBillTotalTx updates the bill total within a transaction
func updateBillTotalTx(ctx context.Context, tx *sqldb.Tx, billID string, newTotal int64) error {
	_, err := tx.Exec(ctx, `
		UPDATE bills SET total_amount = $1 WHERE id = $2
	`, newTotal, billID)
	if err != nil {
		return fmt.Errorf("failed to update bill %s total: %w", billID, err)
	}
	return nil
}

func ListBillsByStatus(
	ctx context.Context,
	status Status,
) ([]*Bill, error) {

	rows, err := db.Query(ctx, `
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
		bill, err := scanBill(rows)
		if err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}

	return bills, nil
}

func ListBillsAll(ctx context.Context) ([]*Bill, error) {
	rows, err := db.Query(ctx, `
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
		bill, err := scanBill(rows)
		if err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}

	return bills, nil
}

func InsertLineItem(ctx context.Context, item *LineItem) error {
	_, err := db.Exec(ctx, `
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
	if err != nil {
		return fmt.Errorf("failed to insert line item %s for bill %s: %w", item.ID, item.BillID, err)
	}
	return nil
}

// InsertLineItemTx inserts a line item within a transaction
func InsertLineItemTx(ctx context.Context, tx *sqldb.Tx, item *LineItem) error {
	_, err := tx.Exec(ctx, `
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
	if err != nil {
		return fmt.Errorf("failed to insert line item %s for bill %s in transaction: %w", item.ID, item.BillID, err)
	}
	return nil
}

func ListLineItems(ctx context.Context, billID string) ([]*LineItem, error) {
	rows, err := db.Query(ctx, `
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

// getLineItemByIDTx retrieves a specific line item by ID and bill ID within a transaction
func getLineItemByIDTx(ctx context.Context, tx *sqldb.Tx, billID, itemID string) (*LineItem, error) {
	row := tx.QueryRow(ctx, `
        SELECT id, amount, currency, description, created_at
        FROM line_items
        WHERE bill_id = $1 AND id = $2
    `, billID, itemID)

	var (
		li          LineItem
		amount      int64
		currencyStr string
	)

	err := row.Scan(
		&li.ID,
		&amount,
		&currencyStr,
		&li.Description,
		&li.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Not found
	}

	if err != nil {
		return nil, err
	}

	li.BillID = billID
	m, err := money.NewMoney(amount, money.Currency(currencyStr))
	if err != nil {
		return nil, err
	}
	li.Amount = m

	return &li, nil
}

// GetLineItemByID retrieves a specific line item by ID and bill ID
func GetLineItemByID(ctx context.Context, billID, itemID string) (*LineItem, error) {
	row := db.QueryRow(ctx, `
        SELECT id, amount, currency, description, created_at
        FROM line_items
        WHERE bill_id = $1 AND id = $2
    `, billID, itemID)

	var (
		li          LineItem
		amount      int64
		currencyStr string
	)

	err := row.Scan(
		&li.ID,
		&amount,
		&currencyStr,
		&li.Description,
		&li.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Not found
	}

	if err != nil {
		return nil, err
	}

	li.BillID = billID
	m, err := money.NewMoney(amount, money.Currency(currencyStr))
	if err != nil {
		return nil, err
	}
	li.Amount = m

	return &li, nil
}
