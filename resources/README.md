# Ledger API

A double-entry ledger API implemented in Go with Gin, PostgreSQL, and GORM. This project follows the [Ledger API Implementation Assignment](Ledger%20API%20Implementation%20Assignment%20(Go,%20Gin,%20Postgres,%20GORM).md), delivered in milestones M1 through M5.

## Project status

The implementation is at the start of M1. Treat the assignment as the source of truth for requirements, endpoint behavior, database constraints, business rules, and acceptance criteria. Build and verify one milestone at a time; do not begin the next milestone until the current one meets its stated checks.

## Technology and conventions

- Go, Gin, PostgreSQL, and GORM.
- Use versioned SQL migrations. Do not use GORM `AutoMigrate`.
- Use UUIDs for account and transfer identifiers, and BIGINT identity keys for ledger entries and jobs.
- Represent money as `int64` minor units. Use `shopspring/decimal` for FX rates; never use floating point for money calculations.
- Store timestamps as UTC `timestamptz` and return RFC 3339 values in JSON.
- Serve application endpoints under `/v1`; expose `/healthz` and `/readyz` outside that path.
- Authenticate `/v1` routes with `X-API-Key` values loaded from `API_KEYS`.
- Return errors using the common `{ "error": { "code": "...", "message": "...", "details": ... } }` envelope described in the assignment.
- Use request IDs and structured `slog` logging.
- Use cursor pagination for list endpoints, with a default limit of 50 and maximum of 200.
- Require `Idempotency-Key` on transfer and deposit creation.

## Milestones

| Milestone | Scope |
| --- | --- |
| M1: Core ledger | Runnable service, health checks, migrations and seeds, API-key and request-ID middleware, accounts, currencies, synchronous same-currency deposits and transfers, idempotency. |
| M2: Async processing and holds | Holds, PostgreSQL job queue, worker pool, retries and recovery, asynchronous deposits and transfers, job operations endpoint. |
| M3: Reversals and reporting | Reversals, immutable ledger entries, account entries and statements, reconciliation endpoint. |
| M4: Multi-currency and FX | Append-only FX rates, rate endpoints, cross-currency posting and exact FX reversals. |
| M5: Provider and polish | External rate refresh, PL/pgSQL posting implementation, comparative benchmark, CI, final OpenAPI and run documentation. |

Follow the assignment’s detailed build sequence and acceptance criteria for each milestone. In particular, finish M1’s validation, integration, idempotency, and concurrency checks before starting M2.

## Core accounting rules

- Every posted transfer creates balanced debit and credit entries per currency.
- A CREDIT increases an account balance; a DEBIT decreases it.
- Customer balances and available balances (`balance - held_amount`) must not become negative.
- Account balance changes and ledger entries are committed atomically. Lock affected accounts in ascending ID order.
- Ledger entries are immutable. Account balances must reconcile to their entries, and active holds must reconcile to `held_amount`.
- Deposits move value from the currency’s FUNDING system account. Customer-facing transfers debit customer accounts; system accounts are used only by internal posting paths.

The assignment’s invariants and error-code table define the full expected behavior.

## Repository layout

The current code is organized under `cmd/` and `internal/`, with packages for configuration, server routing, middleware, accounts, transactions, ledger entries, currencies, and auth. The assignment suggests a target separation between HTTP handlers, domain/service logic, repository interfaces, and database implementations. Evolve the current layout toward that separation as milestones are implemented; services should depend on repository interfaces rather than GORM directly.

Supporting API contract and assignment documents live in this directory:

- [Implementation assignment](Ledger%20API%20Implementation%20Assignment%20(Go,%20Gin,%20Postgres,%20GORM).md) — requirements, schema, endpoint definitions, milestones, invariants, and done criteria.
- [OpenAPI specification](openapi.yaml) — API contract; keep it aligned with the implemented milestone and final API.

## Development workflow

1. Select the next incomplete task from the current milestone in the assignment.
2. Implement it with its required database constraints and error mapping.
3. Add the unit or PostgreSQL integration coverage required for that behavior.
4. Check relevant ledger invariants after tests that move money.
5. Update the OpenAPI contract and this documentation when externally visible behavior changes.
6. Move to the next milestone only after the current milestone’s acceptance checks pass.

The assignment requires real PostgreSQL integration tests, race-enabled concurrency tests, and eventually CI with linting and a coverage threshold. Consult it for the exact test requirements at each stage.
