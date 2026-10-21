-- +goose Up
CREATE TABLE transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type TEXT NOT NULL CHECK (type IN ('TRANSFER', 'DEPOSIT', 'REVERSAL')),
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'POSTED', 'REVERSED', 'FAILED')),
    idempotency_key TEXT NOT NULL UNIQUE,
    request_hash TEXT NOT NULL,
    source_account_id UUID NOT NULL REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    dest_account_id UUID NOT NULL REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    source_amount BIGINT NOT NULL CHECK (source_amount > 0),
    source_currency CHAR(3) NOT NULL,
    dest_amount BIGINT NOT NULL CHECK (dest_amount > 0),
    dest_currency CHAR(3) NOT NULL,
    memo TEXT CHECK (memo IS NULL OR char_length(memo) <= 255),
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    posted_at TIMESTAMPTZ,
    CONSTRAINT chk_transfers_distinct_accounts CHECK (source_account_id <> dest_account_id),
    CONSTRAINT chk_transfers_same_currency_amount CHECK (
        source_currency <> dest_currency OR source_amount = dest_amount
    )
);

CREATE INDEX idx_transfers_source_created_at
    ON transfers (source_account_id, created_at DESC);
CREATE INDEX idx_transfers_dest_created_at
    ON transfers (dest_account_id, created_at DESC);
CREATE INDEX idx_transfers_status ON transfers (status);

-- +goose Down
DROP TABLE transfers;
