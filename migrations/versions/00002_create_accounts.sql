-- +goose Up
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('CUSTOMER', 'SYSTEM')),
    system_role TEXT,
    currency CHAR(3) NOT NULL REFERENCES currencies(code) ON UPDATE CASCADE ON DELETE RESTRICT,
    balance BIGINT NOT NULL DEFAULT 0,
    held_amount BIGINT NOT NULL DEFAULT 0 CHECK (held_amount >= 0),
    status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'FROZEN', 'CLOSED')),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT chk_accounts_system_role CHECK (
        (type = 'SYSTEM' AND system_role IN ('FUNDING', 'FX_CLEARING', 'ROUNDING'))
        OR (type = 'CUSTOMER' AND system_role IS NULL)
    ),
    CONSTRAINT chk_accounts_customer_available_balance CHECK (
        type = 'SYSTEM' OR balance - held_amount >= 0
    )
);

CREATE UNIQUE INDEX ux_accounts_system_role_currency
    ON accounts (system_role, currency)
    WHERE system_role IS NOT NULL;

-- Seed the three system accounts for every supported currency.
INSERT INTO accounts (name, type, system_role, currency, created_at, updated_at)
SELECT role || ' ' || code, 'SYSTEM', role, code, now(), now()
FROM currencies
CROSS JOIN (VALUES ('FUNDING'), ('FX_CLEARING'), ('ROUNDING')) AS system_roles(role);

-- +goose Down
DROP TABLE accounts;
