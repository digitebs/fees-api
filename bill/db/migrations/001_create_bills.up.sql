CREATE TABLE bills (
    id TEXT PRIMARY KEY,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    total_amount BIGINT DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    closed_at TIMESTAMP
);
