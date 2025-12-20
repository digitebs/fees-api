CREATE TABLE line_items (
    id TEXT PRIMARY KEY,
    bill_id TEXT NOT NULL,
    amount BIGINT NOT NULL,
    currency TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL
);
