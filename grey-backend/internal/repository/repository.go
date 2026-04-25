package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/David-Kuku/grey-frontend/grey-backend/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}
func (r *Repository) DB() *sqlx.DB {
	return r.db
}

func (r *Repository) CreateUser(ctx context.Context, tx *sqlx.Tx, email, passwordHash string) (*models.User, error) {
	user := &models.User{}
	err := tx.QueryRowxContext(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2)
		 RETURNING id, email, password_hash, created_at, updated_at`,
		email, passwordHash,
	).StructScan(user)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.GetContext(ctx, user,
		`SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = $1`,
		email,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := r.db.GetContext(ctx, user,
		`SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *Repository) CreateWallet(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	err := tx.QueryRowxContext(ctx,
		`INSERT INTO wallets (user_id) VALUES ($1) RETURNING id, user_id, created_at`,
		userID,
	).StructScan(wallet)
	if err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return wallet, nil
}

func (r *Repository) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	err := r.db.GetContext(ctx, wallet,
		`SELECT id, user_id, created_at FROM wallets WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}
	return wallet, nil
}

func (r *Repository) GetWalletByUserIDTx(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	err := tx.QueryRowxContext(ctx,
		`SELECT id, user_id, created_at FROM wallets WHERE user_id = $1`,
		userID,
	).StructScan(wallet)
	if err != nil {
		return nil, fmt.Errorf("get wallet tx: %w", err)
	}
	return wallet, nil
}

func (r *Repository) GetBalances(ctx context.Context, walletID uuid.UUID) ([]models.WalletBalance, error) {
	var balances []models.WalletBalance
	err := r.db.SelectContext(ctx, &balances,
		`SELECT id, wallet_id, currency, balance, updated_at
		 FROM wallet_balances WHERE wallet_id = $1 ORDER BY currency`,
		walletID,
	)
	if err != nil {
		return nil, fmt.Errorf("get balances: %w", err)
	}
	return balances, nil
}
func (r *Repository) GetBalanceForUpdate(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, currency models.Currency) (*models.WalletBalance, error) {
	bal := &models.WalletBalance{}
	err := tx.QueryRowxContext(ctx,
		`SELECT id, wallet_id, currency, balance, updated_at
		 FROM wallet_balances
		 WHERE wallet_id = $1 AND currency = $2
		 FOR UPDATE`,
		walletID, currency,
	).StructScan(bal)
	if err != nil {
		return nil, fmt.Errorf("get balance for update: %w", err)
	}
	return bal, nil
}
func (r *Repository) UpdateBalance(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, currency models.Currency, newBalance int64) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE wallet_balances SET balance = $1, updated_at = NOW()
		 WHERE wallet_id = $2 AND currency = $3`,
		newBalance, walletID, currency,
	)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}
	return nil
}
func (r *Repository) GetLedgerBalance(ctx context.Context, walletID uuid.UUID, currency models.Currency) (int64, error) {
	var balance sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(signed_amount), 0)
		 FROM ledger_entries
		 WHERE wallet_id = $1 AND currency = $2`,
		walletID, currency,
	).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("get ledger balance: %w", err)
	}
	return balance.Int64, nil
}

func (r *Repository) CreateTransaction(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, txType models.TransactionType, status models.TransactionStatus, currency models.Currency, amount int64, idempotencyKey *string, metadata interface{}) (*models.Transaction, error) {
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}

	txn := &models.Transaction{}
	err = tx.QueryRowxContext(ctx,
		`INSERT INTO transactions (user_id, type, status, currency, amount, idempotency_key, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, user_id, type, status, currency, amount, idempotency_key, metadata, created_at, updated_at`,
		userID, txType, status, currency, amount, idempotencyKey, metaJSON,
	).StructScan(txn)
	if err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}
	return txn, nil
}

func (r *Repository) GetTransactionByIdempotencyKey(ctx context.Context, key string) (*models.Transaction, error) {
	txn := &models.Transaction{}
	err := r.db.GetContext(ctx, txn,
		`SELECT id, user_id, type, status, currency, amount, idempotency_key, metadata, created_at, updated_at
		 FROM transactions WHERE idempotency_key = $1`,
		key,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get transaction by idempotency key: %w", err)
	}
	return txn, nil
}

func (r *Repository) GetTransactionByID(ctx context.Context, userID, txID uuid.UUID) (*models.Transaction, error) {
	txn := &models.Transaction{}
	err := r.db.GetContext(ctx, txn,
		`SELECT id, user_id, type, status, currency, amount, idempotency_key, metadata, created_at, updated_at
		 FROM transactions WHERE id = $1 AND user_id = $2`,
		txID, userID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get transaction by id: %w", err)
	}
	return txn, nil
}

func (r *Repository) UpdateTransactionStatus(ctx context.Context, tx *sqlx.Tx, txID uuid.UUID, status models.TransactionStatus) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, txID,
	)
	if err != nil {
		return fmt.Errorf("update transaction status: %w", err)
	}
	return nil
}

func (r *Repository) CreateLedgerEntry(ctx context.Context, tx *sqlx.Tx, entry *models.LedgerEntry) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO ledger_entries (id, transaction_id, wallet_id, currency, amount, direction, signed_amount, account, description)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		entry.ID, entry.TransactionID, entry.WalletID, entry.Currency,
		entry.Amount, entry.Direction, entry.SignedAmount, entry.Account, entry.Description,
	)
	if err != nil {
		return fmt.Errorf("create ledger entry: %w", err)
	}
	return nil
}

func (r *Repository) GetLedgerEntriesByTransaction(ctx context.Context, txID uuid.UUID) ([]models.LedgerEntry, error) {
	var entries []models.LedgerEntry
	err := r.db.SelectContext(ctx, &entries,
		`SELECT id, transaction_id, wallet_id, currency, amount, direction, signed_amount, account, description, created_at
		 FROM ledger_entries WHERE transaction_id = $1 ORDER BY created_at`,
		txID,
	)
	if err != nil {
		return nil, fmt.Errorf("get ledger entries: %w", err)
	}
	return entries, nil
}

func (r *Repository) CreateFXQuote(ctx context.Context, quote *models.FXQuote) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO fx_quotes (id, user_id, source_currency, target_currency, market_rate, quoted_rate, spread_pct, source_amount, target_amount, fee_amount, fee_currency, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		quote.ID, quote.UserID, quote.SourceCurrency, quote.TargetCurrency,
		quote.MarketRate, quote.QuotedRate, quote.SpreadPct,
		quote.SourceAmount, quote.TargetAmount, quote.FeeAmount, quote.FeeCurrency, quote.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create fx quote: %w", err)
	}
	return nil
}

func (r *Repository) GetFXQuote(ctx context.Context, quoteID uuid.UUID) (*models.FXQuote, error) {
	quote := &models.FXQuote{}
	err := r.db.GetContext(ctx, quote,
		`SELECT id, user_id, source_currency, target_currency, market_rate, quoted_rate, spread_pct,
		        source_amount, target_amount, fee_amount, fee_currency, expires_at, executed, executed_at,
		        transaction_id, created_at
		 FROM fx_quotes WHERE id = $1`,
		quoteID,
	)
	if err != nil {
		return nil, fmt.Errorf("get fx quote: %w", err)
	}
	return quote, nil
}

func (r *Repository) MarkQuoteExecuted(ctx context.Context, tx *sqlx.Tx, quoteID uuid.UUID, txnID uuid.UUID) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE fx_quotes SET executed = TRUE, executed_at = NOW(), transaction_id = $1 WHERE id = $2`,
		txnID, quoteID,
	)
	if err != nil {
		return fmt.Errorf("mark quote executed: %w", err)
	}
	return nil
}

func (r *Repository) GetCachedRate(ctx context.Context, base, target models.Currency) (*float64, error) {
	var rate float64
	err := r.db.QueryRowContext(ctx,
		`SELECT rate FROM fx_rate_cache
		 WHERE base_currency = $1 AND target_currency = $2 AND expires_at > NOW()`,
		base, target,
	).Scan(&rate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get cached rate: %w", err)
	}
	return &rate, nil
}

func (r *Repository) UpsertCachedRate(ctx context.Context, base, target models.Currency, rate float64, expiresAt interface{}) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO fx_rate_cache (base_currency, target_currency, rate, fetched_at, expires_at)
		 VALUES ($1, $2, $3, NOW(), $4)
		 ON CONFLICT (base_currency, target_currency)
		 DO UPDATE SET rate = $3, fetched_at = NOW(), expires_at = $4`,
		base, target, rate, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("upsert cached rate: %w", err)
	}
	return nil
}

func (r *Repository) CreatePayout(ctx context.Context, tx *sqlx.Tx, payout *models.Payout) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO payouts (id, transaction_id, user_id, source_currency, amount, recipient_name, recipient_bank_code, recipient_account, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		payout.ID, payout.TransactionID, payout.UserID, payout.SourceCurrency,
		payout.Amount, payout.RecipientName, payout.RecipientBankCode,
		payout.RecipientAccount, payout.Status,
	)
	if err != nil {
		return fmt.Errorf("create payout: %w", err)
	}
	return nil
}

func (r *Repository) GetPayoutByID(ctx context.Context, payoutID uuid.UUID) (*models.Payout, error) {
	p := &models.Payout{}
	err := r.db.GetContext(ctx, p,
		`SELECT id, transaction_id, user_id, source_currency, amount, recipient_name,
		        recipient_bank_code, recipient_account, status, failure_reason,
		        reversal_transaction_id, created_at, updated_at
		 FROM payouts WHERE id = $1`,
		payoutID,
	)
	if err != nil {
		return nil, fmt.Errorf("get payout: %w", err)
	}
	return p, nil
}

func (r *Repository) GetPayoutByTransactionID(ctx context.Context, txID uuid.UUID) (*models.Payout, error) {
	p := &models.Payout{}
	err := r.db.GetContext(ctx, p,
		`SELECT id, transaction_id, user_id, source_currency, amount, recipient_name,
		        recipient_bank_code, recipient_account, status, failure_reason,
		        reversal_transaction_id, created_at, updated_at
		 FROM payouts WHERE transaction_id = $1`,
		txID,
	)
	if err != nil {
		return nil, fmt.Errorf("get payout by transaction: %w", err)
	}
	return p, nil
}

func (r *Repository) UpdatePayoutStatus(ctx context.Context, tx *sqlx.Tx, payoutID uuid.UUID, status models.PayoutStatus, failureReason *string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE payouts SET status = $1, failure_reason = $2, updated_at = NOW() WHERE id = $3`,
		status, failureReason, payoutID,
	)
	if err != nil {
		return fmt.Errorf("update payout status: %w", err)
	}
	return nil
}

func (r *Repository) SetPayoutReversalTx(ctx context.Context, tx *sqlx.Tx, payoutID uuid.UUID, reversalTxID uuid.UUID) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE payouts SET reversal_transaction_id = $1, updated_at = NOW() WHERE id = $2`,
		reversalTxID, payoutID,
	)
	if err != nil {
		return fmt.Errorf("set payout reversal tx: %w", err)
	}
	return nil
}

func (r *Repository) GetTransactionHistory(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Transaction, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM transactions WHERE user_id = $1`,
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}

	offset := (page - 1) * pageSize
	var txns []models.Transaction
	err = r.db.SelectContext(ctx, &txns,
		`SELECT id, user_id, type, status, currency, amount, idempotency_key, metadata, created_at, updated_at
		 FROM transactions
		 WHERE user_id = $1
		 ORDER BY created_at DESC, id DESC
		 LIMIT $2 OFFSET $3`,
		userID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("get transaction history: %w", err)
	}

	return txns, total, nil
}

func (r *Repository) CreateAuditLog(ctx context.Context, userID *uuid.UUID, action, entityType string, entityID uuid.UUID, requestID string, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		payloadJSON = []byte("{}")
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO audit_log (user_id, action, entity_type, entity_id, request_id, payload)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, action, entityType, entityID, requestID, payloadJSON,
	)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}
