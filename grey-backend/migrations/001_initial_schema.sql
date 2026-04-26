
BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE currency_code AS ENUM ('USD', 'GBP', 'EUR', 'NGN', 'KES');

CREATE TYPE transaction_type AS ENUM ('deposit', 'conversion', 'payout');

CREATE TYPE transaction_status AS ENUM (
    'pending',
    'processing',
    'successful',
    'failed',
    'reversed'
);

CREATE TYPE ledger_direction AS ENUM ('debit', 'credit');

CREATE TYPE payout_status AS ENUM (
    'pending',
    'processing',
    'successful',
    'failed'
);

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);

CREATE TABLE wallets (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL UNIQUE REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE wallet_balances (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id  UUID NOT NULL REFERENCES wallets(id),
    currency   currency_code NOT NULL,
    balance    BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (wallet_id, currency)
);

CREATE TABLE transactions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id),
    type            transaction_type NOT NULL,
    status          transaction_status NOT NULL DEFAULT 'pending',
    currency        currency_code NOT NULL,
    amount          BIGINT NOT NULL CHECK (amount > 0),
    idempotency_key TEXT UNIQUE,
    metadata        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id ON transactions (user_id);
CREATE INDEX idx_transactions_created_at ON transactions (created_at DESC);
CREATE INDEX idx_transactions_idempotency ON transactions (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE TABLE ledger_entries (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    wallet_id      UUID REFERENCES wallets(id),
    currency       currency_code NOT NULL,
    amount         BIGINT NOT NULL CHECK (amount > 0),
    direction      ledger_direction NOT NULL,
    signed_amount  BIGINT NOT NULL,
    account        TEXT NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_entries_transaction ON ledger_entries (transaction_id);
CREATE INDEX idx_ledger_entries_wallet ON ledger_entries (wallet_id)
    WHERE wallet_id IS NOT NULL;
CREATE INDEX idx_ledger_entries_wallet_currency ON ledger_entries (wallet_id, currency)
    WHERE wallet_id IS NOT NULL;

CREATE TABLE fx_quotes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id),
    source_currency currency_code NOT NULL,
    target_currency currency_code NOT NULL,
    market_rate     NUMERIC(20, 10) NOT NULL,
    quoted_rate     NUMERIC(20, 10) NOT NULL,
    spread_pct      NUMERIC(6, 4) NOT NULL,
    source_amount   BIGINT NOT NULL,
    target_amount   BIGINT NOT NULL,
    fee_amount      BIGINT NOT NULL DEFAULT 0,
    fee_currency    currency_code,
    expires_at      TIMESTAMPTZ NOT NULL,
    executed        BOOLEAN NOT NULL DEFAULT FALSE,
    executed_at     TIMESTAMPTZ,
    transaction_id  UUID REFERENCES transactions(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fx_quotes_user ON fx_quotes (user_id);
CREATE INDEX idx_fx_quotes_expires ON fx_quotes (expires_at)
    WHERE executed = FALSE;

CREATE TABLE payouts (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id      UUID NOT NULL UNIQUE REFERENCES transactions(id),
    user_id             UUID NOT NULL REFERENCES users(id),
    source_currency     currency_code NOT NULL,
    amount              BIGINT NOT NULL CHECK (amount > 0),
    recipient_name      TEXT NOT NULL,
    recipient_bank_code TEXT NOT NULL,
    recipient_account   TEXT NOT NULL,
    status              payout_status NOT NULL DEFAULT 'pending',
    failure_reason      TEXT,
    reversal_transaction_id UUID REFERENCES transactions(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payouts_user ON payouts (user_id);
CREATE INDEX idx_payouts_status ON payouts (status)
    WHERE status IN ('pending', 'processing');

CREATE TABLE fx_rate_cache (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    base_currency   currency_code NOT NULL,
    target_currency currency_code NOT NULL,
    rate            NUMERIC(20, 10) NOT NULL,
    fetched_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    UNIQUE (base_currency, target_currency)
);


CREATE OR REPLACE FUNCTION verify_balance(p_wallet_id UUID, p_currency currency_code)
RETURNS BIGINT AS $$
DECLARE
    cached_bal BIGINT;
    ledger_bal BIGINT;
BEGIN
    SELECT COALESCE(balance, 0) INTO cached_bal
    FROM wallet_balances
    WHERE wallet_id = p_wallet_id AND currency = p_currency;

    SELECT COALESCE(SUM(signed_amount), 0) INTO ledger_bal
    FROM ledger_entries
    WHERE wallet_id = p_wallet_id AND currency = p_currency;

    RETURN cached_bal - ledger_bal;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION create_wallet_balances()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO wallet_balances (wallet_id, currency, balance)
    VALUES
        (NEW.id, 'USD', 0),
        (NEW.id, 'GBP', 0),
        (NEW.id, 'EUR', 0),
        (NEW.id, 'NGN', 0),
        (NEW.id, 'KES', 0);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_create_wallet_balances
    AFTER INSERT ON wallets
    FOR EACH ROW
    EXECUTE FUNCTION create_wallet_balances();

COMMIT;
