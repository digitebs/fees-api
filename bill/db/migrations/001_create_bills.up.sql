CREATE TABLE bills (
    id TEXT PRIMARY KEY,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    total_amount BIGINT DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    closed_at TIMESTAMP
);

-- Add indexes for better query performance
CREATE INDEX idx_bills_status ON bills(status);
CREATE INDEX idx_bills_created_at ON bills(created_at DESC);
