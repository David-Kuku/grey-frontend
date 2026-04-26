# Kite — Multi-Currency Wallet

A cross-border payments prototype with a double-entry ledger, FX conversion engine, and simulated payout rails.

**Stack:** React + TypeScript (Vite) · Go (chi) · PostgreSQL 16 · Redis · Docker

---

## Quick Start

```bash
git clone https://github.com/David-Kuku/kuku-kite-app.git
cd kuku-kite-app
docker compose up --build
```

- Frontend: `http://localhost:3000`
- API: `http://localhost:8080`

> `--build` is only needed on the first run or after code changes.

Seed a demo account:

```bash
cd grey-backend && make seed
# Login: demo@kite.test / password123
```

## Running Tests

```bash
docker compose up postgres-test -d
psql "postgres://kite:kite@localhost:5435/kite_test?sslmode=disable" -f grey-backend/migrations/001_initial_schema.sql
cd grey-backend && make test
```

---

## Architecture

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
                      │  └─────────────────────────────────────┘     │
                      └──────────────────────────────────────────────┘
```

**Backend — three layers:**

- **Handlers** — HTTP concerns: parse, validate, call services, respond.
- **Services** — Business logic: Ledger (double-entry), FX (rates/quotes), Payout (state machine + BullMQ worker).
- **Repository** — Data access: all SQL in one place, `sqlx` with parameterised queries.

**Frontend — MVVM:**

- **Pages** — Pure JSX, no logic, just layout and rendering.
- **ViewModels** (`useXxxView` hooks) — All state, mutations, and derived data.
- **Services / Queries** — Axios calls wrapped in TanStack Query hooks.

### Why Go + chi + sqlx

`chi` is a thin stdlib-compatible router. `sqlx` over GORM because for a financial ledger you want to see and control every query — especially the `SELECT ... FOR UPDATE` locks that prevent double-spend.

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
        uuid user_id FK,UK
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
        uuid transaction_id FK,UK
        payout_status status
        text recipient_name
        text recipient_bank_code
        text recipient_account
        uuid reversal_transaction_id FK
    }
```

### Key Design Decisions

**Money as `bigint` in minor units.** Cents for USD/EUR/GBP, kobo for NGN, cents for KES. No floating-point anywhere. `CHECK (balance >= 0)` on `wallet_balances` is a DB-level safety net; real enforcement is via the ledger service under a row lock.

**Double-entry ledger.** Every operation writes ≥2 `ledger_entries` whose `signed_amount` sums to zero per currency. `wallet_balances` is a read-optimised cache rebuildable from `SUM(signed_amount)`.

**FX conversions = two balanced legs.** USD→EUR creates 4 entries in 2 pairs: (1) user USD credit + house USD debit, (2) house EUR credit + user EUR debit. Each pair sums to zero within its currency.

**Concurrency via `SELECT FOR UPDATE`.** Lock the specific `wallet_balances` row before any mutation. Serialises concurrent operations per wallet+currency without table locks.

**Idempotency via unique constraint.** Deposits and payouts require a client `idempotency_key`. Unique constraint on `transactions.idempotency_key` prevents double-processing. Duplicates return the original transaction.

**Payout state machine.** Debited immediately on submission, then `pending → processing → successful|failed`. Failed payouts write inverse ledger entries — append-only, no mutation. Processed via a BullMQ worker (backed by Redis) with exponential backoff retries.

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

## Trade-offs

**Payout worker shares a process with the API.** The background worker that processes payouts runs inside the same Go process as the HTTP server. Simple to run, but means you can't scale one without the other. In production you'd run the worker as a separate service, the queue is already in Redis so no code change is needed, just a different deployment.

**Access tokens last 24 hours.** The current JWT expiry is 24 hours which a long window for a financial app. If a token is stolen, it's valid for up to a day with no way to revoke it. What i would do is to use short-lived access tokens (15 min) paired with a refresh token stored in the database. The short expiry limits the damage window, and revoking the refresh token on logout immediately cuts off access.

**Expired quotes are never deleted.** When a user gets a quote but doesn't confirm it, that row sits in `fx_quotes` forever. It won't cause bugs as expired quotes are rejected on execute but the table grows indefinitely. A nightly cleanup query would fix it; skipped here to keep the system simple.

---

## Scaling to 1M Users

**What breaks first: database connections.**

The API holds a Postgres connection whenever a request interacts with the database, especially during transactions, which causes the connection pool to fill up quickly under load. With Postgres defaulting to around 100 connections, this limit is easily reached as the system scales. At higher scale, like handling millions in traffic, adding more API instances only increases the number of connections and worsens the bottleneck. PgBouncer solves this by sitting between the API and Postgres and reusing a small pool of database connections efficiently, preventing exhaustion while supporting high traffic.

Fix: put PgBouncer in front of Postgres in transaction-pooling mode. Connections are returned to the pool between statements, so hundreds of API instances share a small pool of real Postgres connections.

**What breaks second: FX rate cache stampede.**

When the Redis cache expires, concurrent quote requests all see a cache miss simultaneously and pile onto the Frankfurter API — the same rate fetched hundreds of times at once, risking rate-limit errors.

Fix: wrap the fetch in `singleflight` (`golang.org/x/sync/singleflight`). Only one goroutine fetches; the rest block and share the result. One external call per cache miss regardless of concurrency.

---

## Bonus Features

- **Observability** — Structured JSON logging via `slog`, request ID on every log line and response header.
- **Per-user rate limiting** — Token bucket (via `golang.org/x/time/rate`) scoped per user ID, not IP. Separate limits for conversions (5 rps / burst 10) and payouts (3 rps / burst 5).

---

## Project Structure

```
grey-frontend/                         # Monorepo root
├── docker-compose.yml                 # Spins up postgres, redis, api, frontend
├── frontend/                          # React + Vite
│   ├── Dockerfile
│   ├── nginx.conf                     # Proxies /api/ to the api container
│   └── src/
│       ├── components/                # Shared UI components
│       ├── queries/                   # TanStack Query hooks
│       ├── services/                  # Axios API clients
│       ├── store/                     # Zustand auth store
│       ├── types/                     # Shared TypeScript types
│       ├── utils/                     # Currency, date, idempotency
│       └── views/
│           ├── pages/                 # Pure JSX page components
│           └── viewmodel/             # Hooks with all page logic (MVVM)
└── grey-backend/                      # Go API
    ├── cmd/api/main.go                # Entry point
    ├── internal/
    │   ├── auth/                      # JWT + bcrypt
    │   ├── config/                    # Env-based config
    │   ├── fx/                        # FX rates, caching, quoting
    │   ├── handlers/                  # HTTP handlers (one per domain)
    │   ├── ledger/                    # Double-entry ledger service
    │   ├── middleware/                # Auth, request ID, rate limiting
    │   ├── models/                    # Domain types + DTOs
    │   ├── payout/                    # BullMQ worker + state machine
    │   ├── repository/                # All DB operations (sqlx)
    │   └── server/                    # Router + DI wiring
    ├── migrations/                    # SQL schema
    ├── tests/                         # Integration tests
    ├── Dockerfile
    └── Makefile
```

## Loom Walkthrough

> [TODO: Link]

## Time Spent

~X hours (TODO)
