CREATE TABLE line_items (
    id TEXT PRIMARY KEY,
    bill_id TEXT NOT NULL,
    amount BIGINT NOT NULL,
    currency TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL
);

-- Add indexes for better query performance
CREATE INDEX idx_line_items_bill_id ON line_items(bill_id);
CREATE INDEX idx_line_items_created_at ON line_items(created_at ASC);
