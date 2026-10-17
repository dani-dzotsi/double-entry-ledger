# Ledger API: Implementation Assignment

Go (Gin) · PostgreSQL · GORM · Milestones M1 to M5

## 0. How to use this assignment

Work milestone by milestone. Each milestone lists what to build, which endpoints it touches, and the tests that must pass before you move on. Do not start M2 until M1 is green.

## 1. Global conventions

- **Base path:** `/v1`. Health endpoints sit outside it.
- **Auth:** middleware checks an `X-API-Key` header against keys loaded from the `API_KEYS` env var (comma-separated). Missing or invalid gives 401. Health endpoints are unauthenticated.
- **IDs:** UUID generated in Go (v7 if available, else v4). `ledger_entries` and `jobs` use BIGINT identity instead.
- **Money:** `int64` minor units everywhere, never float. FX rates use `shopspring/decimal`, stored as `NUMERIC(18,8)`.
- **Timestamps:** `timestamptz`, UTC, RFC 3339 in JSON.
- **Pagination:** `?limit` (default 50, max 200) and `?cursor` (opaque, encodes created\_at + id). Responses return `items` and `next_cursor`.
- **Idempotency:** `POST /transfers` and `POST /deposits` require an `Idempotency-Key` header. Same key with the same body returns the original result. Same key with a different body returns 409.
- **Errors:** every error uses one envelope: `error.code`, `error.message`, `error.details` (optional).
- **Logging:** `slog`, with a `request_id` on every line.

### Error codes

| Code | HTTP | When |
| --- | --- | --- |
| VALIDATION\_ERROR | 400 | Bad body, bad query, amount <= 0, unknown currency |
| UNAUTHORIZED | 401 | Missing or invalid API key |
| NOT\_FOUND | 404 | Account, transfer or rate does not exist |
| IDEMPOTENCY\_CONFLICT | 409 | Same key, different request body |
| INVALID\_STATE | 409 | Action not allowed in current status (e.g. reversing a non-POSTED transfer) |
| INSUFFICIENT\_FUNDS | 422 | Available balance too low |
| ACCOUNT\_NOT\_ACTIVE | 422 | Source or destination is FROZEN or CLOSED |
| CURRENCY\_MISMATCH | 422 | Amount currency does not match the source account (M1 to M3: also source and destination differ) |
| FX\_RATE\_UNAVAILABLE | 422 | No usable rate for the currency pair |
| INTERNAL | 500 | Anything unexpected (never leak internals) |

## 2. Data models

Create these tables with SQL migrations (goose or golang-migrate), then map them with GORM structs. Do not use AutoMigrate.

### currencies

| Column | Type | Constraints |
| --- | --- | --- |
| code | CHAR(3) | PK, uppercase |
| name | TEXT | NOT NULL |
| exponent | SMALLINT | NOT NULL, CHECK 0..6 (2 for USD/GHS/EUR, 0 for JPY) |
| active | BOOLEAN | NOT NULL DEFAULT true |

### accounts

| Column | Type | Constraints |
| --- | --- | --- |
| id | UUID | PK |
| name | TEXT | NOT NULL |
| type | TEXT | NOT NULL, CHECK IN (CUSTOMER, SYSTEM) |
| system\_role | TEXT | NULL for customers; for SYSTEM: FUNDING, FX\_CLEARING, ROUNDING |
| currency | CHAR(3) | NOT NULL, FK currencies.code |
| balance | BIGINT | NOT NULL DEFAULT 0 (posted balance) |
| held\_amount | BIGINT | NOT NULL DEFAULT 0, CHECK >= 0 |
| status | TEXT | NOT NULL DEFAULT ACTIVE, CHECK IN (ACTIVE, FROZEN, CLOSED) |
| created\_at, updated\_at | TIMESTAMPTZ | NOT NULL |

Extra constraints:

- `CHECK (type = 'SYSTEM' OR balance - held_amount >= 0)` so customer accounts can never go negative, while system accounts are exempt.
- Unique index on `(system_role, currency)` where `system_role IS NOT NULL`.
- Available balance is derived: `balance - held_amount`.

### transfers

| Column | Type | Constraints |
| --- | --- | --- |
| id | UUID | PK |
| type | TEXT | NOT NULL, CHECK IN (TRANSFER, DEPOSIT, REVERSAL) |
| status | TEXT | NOT NULL, CHECK IN (PENDING, POSTED, REVERSED, FAILED) |
| idempotency\_key | TEXT | NOT NULL, UNIQUE |
| request\_hash | TEXT | NOT NULL (hash of normalized body, for conflict detection) |
| source\_account\_id | UUID | NOT NULL, FK accounts |
| dest\_account\_id | UUID | NOT NULL, FK accounts, CHECK <> source |
| source\_amount | BIGINT | NOT NULL, CHECK > 0 |
| source\_currency | CHAR(3) | NOT NULL |
| dest\_amount | BIGINT | NOT NULL, CHECK > 0 (equals source\_amount when same currency) |
| dest\_currency | CHAR(3) | NOT NULL |
| fx\_rate\_id | UUID | NULL, FK fx\_rates (set for cross-currency) |
| fx\_rounding\_residue | NUMERIC(18,8) | NULL (sub-minor-unit remainder, informational) |
| reversal\_of\_id | UUID | NULL, FK transfers, UNIQUE (a transfer can be reversed once) |
| memo | TEXT | NULL, max 255 |
| failure\_reason | TEXT | NULL (an error code from section 1) |
| created\_at, updated\_at | TIMESTAMPTZ | NOT NULL |
| posted\_at | TIMESTAMPTZ | NULL |

Indexes: `(source_account_id, created_at DESC)`, `(dest_account_id, created_at DESC)`, `(status)`.

### ledger\_entries (immutable)

| Column | Type | Constraints |
| --- | --- | --- |
| id | BIGINT | identity PK |
| transfer\_id | UUID | NOT NULL, FK transfers |
| account\_id | UUID | NOT NULL, FK accounts |
| direction | TEXT | NOT NULL, CHECK IN (DEBIT, CREDIT) |
| amount | BIGINT | NOT NULL, CHECK > 0 |
| currency | CHAR(3) | NOT NULL |
| balance\_after | BIGINT | NOT NULL (account balance after this entry) |
| created\_at | TIMESTAMPTZ | NOT NULL |

Index: `(account_id, id DESC)` for statements. Add a trigger (M3) that raises on UPDATE or DELETE.

Direction semantics: for any account, CREDIT increases `balance` and DEBIT decreases it. Every transfer must produce entries where, per currency, total debits equal total credits.

### holds

| Column | Type | Constraints |
| --- | --- | --- |
| id | UUID | PK |
| transfer\_id | UUID | NOT NULL, UNIQUE, FK transfers |
| account\_id | UUID | NOT NULL, FK accounts |
| amount | BIGINT | NOT NULL, CHECK > 0 |
| status | TEXT | NOT NULL, CHECK IN (ACTIVE, CAPTURED, RELEASED) |
| created\_at | TIMESTAMPTZ | NOT NULL |
| resolved\_at | TIMESTAMPTZ | NULL |

M2 and later only.

### fx\_rates

| Column | Type | Constraints |
| --- | --- | --- |
| id | UUID | PK |
| base\_currency | CHAR(3) | NOT NULL |
| quote\_currency | CHAR(3) | NOT NULL, CHECK <> base |
| rate | NUMERIC(18,8) | NOT NULL, CHECK > 0 (1 base = rate quote) |
| source | TEXT | NOT NULL, CHECK IN (MANUAL, PROVIDER) |
| effective\_at | TIMESTAMPTZ | NOT NULL |
| created\_at | TIMESTAMPTZ | NOT NULL |

Index: `(base_currency, quote_currency, effective_at DESC)`. Rates are append-only: a new rate never edits an old one, because transfers reference the exact row used. M4 and later.

### jobs (the Postgres queue)

| Column | Type | Constraints |
| --- | --- | --- |
| id | BIGINT | identity PK |
| type | TEXT | NOT NULL (e.g. POST\_TRANSFER) |
| payload | JSONB | NOT NULL (e.g. transfer\_id) |
| status | TEXT | NOT NULL, CHECK IN (QUEUED, RUNNING, DONE, FAILED) |
| attempts | INT | NOT NULL DEFAULT 0 |
| max\_attempts | INT | NOT NULL DEFAULT 5 |
| run\_at | TIMESTAMPTZ | NOT NULL (supports retry backoff) |
| locked\_at | TIMESTAMPTZ | NULL |
| last\_error | TEXT | NULL |
| created\_at, updated\_at | TIMESTAMPTZ | NOT NULL |

Index: `(status, run_at)`. M2 and later.

### Seed data (migration or startup seed)

- Currencies: USD (2), GHS (2), EUR (2), JPY (0).
- For each currency, three SYSTEM accounts: FUNDING, FX\_CLEARING, ROUNDING.

## 3. Endpoints

### Health

| Method | Path | Milestone | Purpose |
| --- | --- | --- | --- |
| GET | /healthz | M1 | Liveness. Always 200 if the process is up |
| GET | /readyz | M1 | Readiness. 200 only if the DB ping succeeds, else 503 |

### Currencies

| Method | Path | Milestone | Purpose |
| --- | --- | --- | --- |
| GET | /v1/currencies | M1 | List active currencies with exponents |

### Accounts

| Method | Path | Milestone | Purpose | Success |
| --- | --- | --- | --- | --- |
| POST | /v1/accounts | M1 | Create a CUSTOMER account (name, currency) | 201 |
| GET | /v1/accounts | M1 | List accounts, filter by `currency`, `status`, `type`; paginated | 200 |
| GET | /v1/accounts/:id | M1 | Account with balance, held\_amount, available\_balance | 200 |
| PATCH | /v1/accounts/:id/status | M1 | Set ACTIVE, FROZEN or CLOSED (CLOSED only if balance and holds are 0) | 200 |
| GET | /v1/accounts/:id/entries | M3 | Ledger entries for the account, newest first, paginated | 200 |
| GET | /v1/accounts/:id/statement | M3 | Entries between `from` and `to`, with opening balance, closing balance and totals | 200 |

SYSTEM accounts are created only by seed, never through the API. They are readable through GET.

### Deposits

| Method | Path | Milestone | Purpose | Success |
| --- | --- | --- | --- | --- |
| POST | /v1/deposits | M1 | Fund a customer account: body has `account_id`, `amount`, `currency`, optional `memo`. Internally a transfer of type DEPOSIT from the FUNDING system account | 201 (M1), 202 (M2+) |

### Transfers

| Method | Path | Milestone | Purpose | Success |
| --- | --- | --- | --- | --- |
| POST | /v1/transfers | M1 | Create a transfer. M1 to M2 launch: same-currency, posted synchronously. From M2: asynchronous | 201 (M1), 202 (M2+) |
| GET | /v1/transfers/:id | M1 | Transfer with status, amounts, FX info, failure\_reason | 200 |
| GET | /v1/transfers | M1 | List, filter by `account_id`, `status`, `type`, date range; paginated | 200 |
| POST | /v1/transfers/:id/reversal | M3 | Reverse a POSTED transfer. Returns the new REVERSAL transfer | 202 |

`POST /v1/transfers` body fields:

| Field | Type | Rules |
| --- | --- | --- |
| source\_account\_id | UUID | required, must be CUSTOMER and ACTIVE |
| dest\_account\_id | UUID | required, ACTIVE, different from source |
| amount | int64 | required, > 0, in source account currency minor units |
| currency | string | required, must equal source account currency |
| memo | string | optional, max 255 |

Response fields: `id`, `type`, `status`, `source_account_id`, `dest_account_id`, `source_amount`, `source_currency`, `dest_amount`, `dest_currency`, `fx_rate` (object or null), `memo`, `failure_reason`, `created_at`, `posted_at`.

From M4, if source and dest currencies differ, the dest amount is computed from the latest rate. From M2, clients poll `GET /v1/transfers/:id` until status is POSTED or FAILED. Customer-facing transfers may only debit CUSTOMER accounts. Only the deposit path and internal FX posting touch SYSTEM accounts.

### FX rates

| Method | Path | Milestone | Purpose |
| --- | --- | --- | --- |
| POST | /v1/fx-rates | M4 | Add a MANUAL rate (base, quote, rate, optional effective\_at) |
| GET | /v1/fx-rates/latest | M4 | Latest rate for `base` and `quote` |
| GET | /v1/fx-rates | M4 | History for a pair, paginated |
| POST | /v1/fx-rates/refresh | M5 | Pull rates from the external provider and store them as PROVIDER rows |

### Operations

| Method | Path | Milestone | Purpose |
| --- | --- | --- | --- |
| GET | /v1/admin/reconciliation | M3 | Run invariant checks (see section 5) and return pass/fail per check with offending ids |
| GET | /v1/admin/jobs | M2 | Counts of jobs by status, oldest QUEUED age, recent failures |

## 4. Milestones

### M1: Core ledger (synchronous)

**Build**

1. Project skeleton, config from env, `slog`, graceful shutdown, `/healthz` and `/readyz`.
2. Migrations for currencies, accounts, transfers, ledger\_entries. Seed data.
3. Repository layer (interfaces first, GORM implementations second). Services never import GORM.
4. API-key middleware, request ID middleware, error-envelope middleware.
5. Endpoints: currencies, accounts (create, list, get, status), deposits, transfers (create, get, list).
6. Idempotency using `idempotency_key` plus `request_hash`.

**Posting algorithm (the core of the project)**

1. Begin a transaction.
2. Lock both account rows with `SELECT ... FOR UPDATE`, always in ascending id order to avoid deadlocks (GORM: `clause.Locking{Strength: "UPDATE"}` with `Order("id")`).
3. Validate: both ACTIVE, currency matches, available balance covers the amount (customer sources).
4. Insert the transfer (status POSTED, posted\_at set).
5. Insert a DEBIT entry for the source and a CREDIT entry for the destination, each with `balance_after`.
6. Update both balances.
7. Commit. On any failure, roll back and return the mapped error.

**Must pass before M2:** unit tests for validation, integration tests (testcontainers Postgres) for each endpoint, idempotency replay and conflict, and a concurrency test (100 goroutines transferring between the same accounts, final balances correct, no negative balance, no deadlock).

### M2: Async processing and holds

**Build**

1. Migrations for holds and jobs.
2. `POST /transfers` and `POST /deposits` now run one transaction that: locks the source account, checks available balance (insufficient gives synchronous 422), inserts the transfer as PENDING, inserts an ACTIVE hold, increases `held_amount`, inserts a POST\_TRANSFER job, then returns 202.
3. Worker pool (configurable size) claiming jobs with `SELECT ... FOR UPDATE SKIP LOCKED` ordered by `run_at`, inside a transaction that also marks the job RUNNING.
4. Worker posting step: re-validate (accounts still ACTIVE), capture the hold, post entries as in M1, set the transfer POSTED. If it cannot post for a permanent reason, set the transfer FAILED with `failure_reason`, release the hold, restore `held_amount`.
5. Retries with exponential backoff for transient errors, up to `max_attempts`, then FAILED. A stale RUNNING job (locked\_at too old) is reclaimed.
6. `GET /v1/admin/jobs`.

**Must pass before M3:** crash-recovery test (kill the worker mid-job, job is reclaimed and the transfer posts exactly once), duplicate-claim test (many workers, each job processed once), hold accounting test (held\_amount always equals the sum of ACTIVE holds), and the M1 concurrency test again with the async path.

### M3: Reversals, statements, reconciliation

**Build**

1. `POST /v1/transfers/:id/reversal`: only for POSTED transfers, only once (enforced by the unique `reversal_of_id`). It creates a REVERSAL transfer that flows through the same hold and worker path, with source and destination swapped and the original amounts. When it posts, the original becomes REVERSED.
2. If the original destination can no longer afford the reversal, the REVERSAL transfer ends FAILED with INSUFFICIENT\_FUNDS and the original stays POSTED.
3. `GET /accounts/:id/entries` and `/statement` with cursor pagination and correct opening and closing balances.
4. Immutability trigger on `ledger_entries`.
5. `GET /v1/admin/reconciliation`.

**Must pass before M4:** property-style test (random sequences of transfers and reversals, then all invariants hold), statement test (opening + credits - debits = closing), reversal edge cases (twice, non-POSTED, insufficient funds).

### M4: Multi-currency with FX

**Build**

1. Migration for fx\_rates and the extra transfer columns already listed.
2. Manual rate endpoints.
3. Cross-currency transfers: the source account debits in its currency, the destination credits in its currency. Compute `dest_amount = round_half_even(source_amount * rate * 10^(dest_exp - source_exp))`. Store the `fx_rate_id` used, and store the sub-unit remainder in `fx_rounding_residue`.
4. Four-entry posting through the FX\_CLEARING accounts: DEBIT source account, CREDIT FX\_CLEARING of the source currency, DEBIT FX\_CLEARING of the destination currency, CREDIT destination account. Debits equal credits within each currency.
5. Rate selection: the latest rate with `effective_at <= now` for the pair, or the inverse of the opposite pair if only that exists. If none, FX\_RATE\_UNAVAILABLE.
6. Reversals of cross-currency transfers reuse the original dest amount and rate exactly. Never recompute.

**Must pass before M5:** rounding tests (half-even boundaries, JPY exponent 0 vs USD exponent 2), a per-currency balance test, and reversal exactness.

### M5: External rates, PL/pgSQL comparison, polish

**Build**

1. A rates provider interface with one HTTP implementation, a timeout, retries, and a fake for tests. `POST /v1/fx-rates/refresh` stores PROVIDER rows.
2. Rewrite transfer posting as a PL/pgSQL function behind the same repository interface. The same test suite must pass against both implementations.
3. Write a benchmark comparing Go-driven posting with the function (transfers per second at 1, 10 and 50 concurrent workers), and record the results in the README.
4. Polish: CI with a coverage gate, OpenAPI file matching the final API, README with run instructions and architecture diagram, Makefile targets.

## 5. Business rules and invariants

These are what `/v1/admin/reconciliation` and your tests must verify.

1. Per transfer, per currency: sum of DEBIT amounts equals sum of CREDIT amounts.
2. For every account: `balance` equals the sum of its CREDIT entries minus its DEBIT entries.
3. For every account: `balance_after` of the latest entry equals `balance`.
4. For every CUSTOMER account: `balance >= 0` and `balance - held_amount >= 0`.
5. For every account: `held_amount` equals the sum of its ACTIVE holds.
6. Per currency, the sum of all account balances is 0 once the FUNDING account (which goes negative as deposits are made) is included.
7. Every POSTED transfer has entries. No PENDING or FAILED transfer has entries.
8. A transfer has at most one reversal. A REVERSED transfer has exactly one POSTED REVERSAL.
9. Ledger entries are never updated or deleted.
10. An idempotency key maps to exactly one transfer.

## 6. Suggested project layout

```
cmd/api/main.go
cmd/worker/main.go            (or run the worker inside the API process behind a flag)
internal/config/
internal/http/                (gin router, handlers, middleware, error mapping)
internal/domain/              (types, status enums, errors, pure rules)
internal/service/             (accounts, transfers, fx, reconciliation)
internal/repository/          (interfaces)
internal/repository/gormrepo/ (GORM implementations)
internal/repository/pgfunc/   (M5 PL/pgSQL implementation)
internal/worker/
internal/fxprovider/
migrations/
docs/openapi.yaml
```

Rule: handlers call services, services call repository interfaces, only `gormrepo` and `pgfunc` import GORM or SQL.

## 7. Testing requirements (across all milestones)

- Unit tests for domain rules and FX rounding.
- Integration tests against a real Postgres via testcontainers, one fresh schema per test package.
- Concurrency tests with the `-race` flag.
- Invariant checks run at the end of every integration test that moves money.
- CI runs lint, tests with `-race`, and enforces a coverage threshold.

## 8. Deliverables and definition of done

- All endpoints in section 3 working with the documented status codes and error codes.
- All tables created by versioned migrations, with every constraint in section 2.
- All invariants in section 5 verified by tests and by the reconciliation endpoint.
- OpenAPI file, README, and the benchmark results.
- Graceful shutdown that stops accepting requests and lets in-flight jobs finish.
