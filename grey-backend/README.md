# Kite — Multi-Currency Wallet Service

A cross-border payments prototype with a double-entry ledger, FX conversion engine, and simulated payout rails.

## Quick Start

```bash
git clone https://github.com/David-Kuku/kite/kite-backend.git
cd kite
docker compose up --build
```

API: `http://localhost:8080` | Frontend: `http://localhost:3000`

Seed a demo account:

```bash
make seed
# Login: demo@kite.test / password123
```

## Running Without Docker

```bash
# 1. Start Postgres
# 2. Run migration
psql "postgres://kite:kite@localhost:5432/kite?sslmode=disable" -f migrations/001_initial_schema.sql

# 3. Configure
cp .env.example .env

# 4. Run
go run ./cmd/api
```

## Running Tests

```bash
docker compose up postgres-test -d
psql "postgres://kite:kite@localhost:5433/kite_test?sslmode=disable" -f migrations/001_initial_schema.sql
make test        # or: make test-race
```

---

## Architecture Overview

```
┌─────────────┐       ┌──────────────────────────────────────────────┐
│   React UI  │◄─────►│                Go API (chi)                  │
│  (Vite +    │  JSON  │                                              │
│  TanStack)  │       │  ┌──────────┐  ┌──────────┐  ┌───────────┐  │
└─────────────┘       │  │ Handlers │──│ Services  │──│   Repo    │  │
                      │  │ (HTTP)   │  │ (Business)│  │  (sqlx)   │  │
                      │  └──────────┘  └──────────┘  └─────┬─────┘  │
                      │                                     │        │
                      │  ┌──────────────────────────────────┘        │
                      │  ▼                                           │
                      │  ┌─────────────────────────────────────┐     │
                      │  │           PostgreSQL 16              │     │
                      │  │  users ── wallets ── wallet_balances │     │
                      │  │       transactions ── ledger_entries │     │
                      │  │  fx_quotes  payouts  fx_rate_cache  │     │
                      │  │           audit_log                 │     │
                      │  └─────────────────────────────────────┘     │
                      └──────────────────────────────────────────────┘
```

Three-layer architecture:

- **Handlers** — HTTP concerns: parse, validate, call services, respond.
- **Services** — Business logic: Ledger (double-entry), FX (rates/quotes), Payout (state machine).
- **Repository** — Data access: all SQL in one place, `sqlx` with parameterised queries.

### Why Go + chi + sqlx

`chi` is a thin stdlib-compatible router. `sqlx` over GORM/ent because for a financial ledger you want to see and control every query — especially the `SELECT ... FOR UPDATE` locks that prevent double-spend.

### Why JWT

Stateless and horizontally scalable without Redis for sessions. In production: short-lived access tokens (15 min) + refresh tokens + revocation list in Redis.

---

## Data Model

```mermaid
erDiagram
    users ||--|| wallets : "has one"
    wallets ||--o{ wallet_balances : "one per currency"
    users ||--o{ transactions : "creates"
    transactions ||--o{ ledger_entries : "produces"
    wallets ||--o{ ledger_entries : "affects"
    users ||--o{ fx_quotes : "requests"
    fx_quotes |o--|| transactions : "executed via"
    transactions ||--o| payouts : "tracks"

    users {
        uuid id PK
        text email UK
        text password_hash
    }
    wallets {
        uuid id PK
        uuid user_id FK_UK
    }
    wallet_balances {
        uuid wallet_id FK
        currency_code currency
        bigint balance "minor units >= 0"
    }
    transactions {
        uuid id PK
        uuid user_id FK
        transaction_type type
        transaction_status status
        text idempotency_key UK
        jsonb metadata
    }
    ledger_entries {
        uuid id PK
        uuid transaction_id FK
        uuid wallet_id FK "nullable for system accts"
        currency_code currency
        bigint amount "always positive"
        ledger_direction direction
        bigint signed_amount "+debit / -credit"
        text account
    }
    fx_quotes {
        uuid id PK
        numeric market_rate
        numeric quoted_rate
        bigint source_amount
        bigint target_amount
        timestamptz expires_at
        boolean executed
    }
    payouts {
        uuid id PK
        uuid transaction_id FK_UK
        payout_status status
        text recipient_name
        text recipient_bank_code
        text recipient_account
        uuid reversal_transaction_id FK
    }
```

### Key Design Decisions

**Money as `bigint` in minor units.** Cents for USD/EUR/GBP, kobo for NGN, cents for KES. No floating-point anywhere. `CHECK (balance >= 0)` on `wallet_balances` is a DB-level safety net; real enforcement is via the ledger service under a row lock.

**Double-entry ledger.** Every operation writes ≥2 `ledger_entries` whose `signed_amount` sums to zero per currency. `wallet_balances` is a cache rebuildable from `SUM(signed_amount)`. The `verify_balance()` SQL function validates this.

**FX conversions = two balanced legs.** USD→EUR creates 4 entries in 2 pairs: (1) user USD credit + house USD debit, (2) house EUR credit + user EUR debit. Each pair sums to zero within its currency.

**Concurrency via `SELECT FOR UPDATE`.** Lock the specific `wallet_balances` row before any mutation. Serialises concurrent operations per wallet+currency without table locks.

**Idempotency via unique constraint.** Deposits and payouts require a client `idempotency_key`. Unique constraint on `transactions.idempotency_key` prevents double-processing. Duplicates return the original transaction.

**Payout state machine.** Debited immediately, then `pending → processing → successful|failed`. Failed payouts write inverse ledger entries — append-only, no mutation.

---

## API Endpoints

| Method | Path                          | Auth | Description           |
| ------ | ----------------------------- | ---- | --------------------- |
| POST   | `/api/v1/auth/signup`         | No   | Create account        |
| POST   | `/api/v1/auth/login`          | No   | Get JWT               |
| GET    | `/api/v1/wallet/balances`     | Yes  | All currency balances |
| POST   | `/api/v1/deposits`            | Yes  | Simulated deposit     |
| POST   | `/api/v1/conversions/quote`   | Yes  | Get FX quote          |
| POST   | `/api/v1/conversions/execute` | Yes  | Execute quote         |
| POST   | `/api/v1/payouts`             | Yes  | Initiate payout       |
| GET    | `/api/v1/transactions`        | Yes  | Paginated history     |
| GET    | `/api/v1/health`              | No   | Health check          |

### Error Format

```json
{
  "code": "INSUFFICIENT_BALANCE",
  "message": "Insufficient balance in NGN. Available: NGN 500.00",
  "details": { "amount": "must be positive" }
}
```

Codes: `VALIDATION_ERROR`, `INVALID_CREDENTIALS`, `EMAIL_EXISTS`, `INSUFFICIENT_BALANCE`, `QUOTE_EXPIRED`, `QUOTE_ALREADY_EXECUTED`, `QUOTE_NOT_FOUND`, `AUTH_REQUIRED`, `INVALID_TOKEN`, `INTERNAL_ERROR`.

---

## Tests

| Test                          | What it proves                                                   |
| ----------------------------- | ---------------------------------------------------------------- |
| `TestLedgerBalancesCorrectly` | Multiple deposits sum correctly across currencies                |
| `TestDepositIdempotency`      | Same key → same transaction, balance moves once                  |
| `TestConcurrentConversions`   | 5 goroutines race; at most 1 succeeds, balance never negative    |
| `TestExpiredQuoteRejection`   | Expired quote → 410 Gone + `QUOTE_EXPIRED`                       |
| `TestFailedPayoutReversal`    | Failure triggers reversal, balance restored                      |
| `TestLedgerReconciliation`    | Cached balance == `SUM(signed_amount)` for every wallet+currency |
| `TestInsufficientBalance`     | Payout on zero balance → 422                                     |

---

## Trade-offs

**FX rates:** Frankfurter (free, no key). Production: paid provider + Redis cache with single-flight pattern.

**Payout simulation:** `time.AfterFunc` goroutine — lost on restart. Production: Temporal/Asynq/River with persistent jobs.

**No migration tool:** Raw `.sql`. Production: golang-migrate or Atlas.

**Quote expiry:** Checked on execute only. Production: background job to clean expired quotes.

**Single process:** No Redis, no queue, no worker separation. First thing to split at scale.

---

## Scaling to 1M Users

1. **DB connections** — PgBouncer + read replicas for balance reads / transaction history.
2. **Hot row on house account** — Shard the house FX pool account or use event sourcing.
3. **FX cache stampede** — Redis + singleflight so one goroutine fetches, others wait.
4. **Payout workers** — Dedicated pool consuming from SQS/Redis streams.
5. **Transaction history** — Cursor pagination (keyset on `(created_at, id)`), table partitioning.
6. **Audit log** — Partition by month, archive to S3 after 90 days.

---

## Bonus Features

- **Audit log** — Immutable `audit_log` table with request IDs threading through the lifecycle.
- **Observability** — Structured JSON logging via `slog`, request ID middleware on every log line and response header.

---

## Project Structure

```
kite/
├── cmd/api/main.go              # Entry point
├── internal/
│   ├── auth/                    # JWT + bcrypt
│   ├── config/                  # Env-based config
│   ├── fx/                      # FX rates, caching, quoting
│   ├── handlers/                # HTTP handlers (one per domain)
│   ├── ledger/                  # Double-entry ledger service
│   ├── middleware/              # Auth, request ID, logging
│   ├── models/                  # Domain types + DTOs
│   ├── payout/                  # State machine + simulation
│   ├── repository/             # All DB operations
│   └── server/                  # Router + DI wiring
├── migrations/                  # SQL schema
├── tests/                       # Integration tests
├── frontend/                    # React + TanStack Query
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```

## Loom Walkthrough

> [TODO: Link]

## Time Spent

~X hours (TODO)
