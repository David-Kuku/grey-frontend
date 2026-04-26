package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/ledger"
	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/models"
	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func connectTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://kite:kite@localhost:5435/kite_test?sslmode=disable"
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func seedUser(t *testing.T, db *sqlx.DB) (userID uuid.UUID, walletID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	repo := repository.New(db)

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatalf("seedUser begin tx: %v", err)
	}
	user, err := repo.CreateUser(ctx, tx, fmt.Sprintf("test-%s@kite.test", uuid.New()), "testhash")
	if err != nil {
		tx.Rollback()
		t.Fatalf("seedUser create user: %v", err)
	}
	wallet, err := repo.CreateWallet(ctx, tx, user.ID)
	if err != nil {
		tx.Rollback()
		t.Fatalf("seedUser create wallet: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("seedUser commit: %v", err)
	}

	uID, wID := user.ID, wallet.ID
	t.Cleanup(func() {
		db.ExecContext(ctx, `DELETE FROM payouts WHERE user_id = $1`, uID)
		db.ExecContext(ctx, `DELETE FROM ledger_entries WHERE transaction_id IN (SELECT id FROM transactions WHERE user_id = $1)`, uID)
		db.ExecContext(ctx, `DELETE FROM fx_quotes WHERE user_id = $1`, uID)
		db.ExecContext(ctx, `DELETE FROM transactions WHERE user_id = $1`, uID)
		db.ExecContext(ctx, `DELETE FROM wallet_balances WHERE wallet_id = $1`, wID)
		db.ExecContext(ctx, `DELETE FROM wallets WHERE user_id = $1`, uID)
		db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, uID)
	})

	return uID, wID
}

func getBalance(t *testing.T, repo *repository.Repository, walletID uuid.UUID, currency models.Currency) int64 {
	t.Helper()
	balances, err := repo.GetBalances(context.Background(), walletID)
	if err != nil {
		t.Fatalf("getBalance: %v", err)
	}
	for _, b := range balances {
		if b.Currency == currency {
			return b.Balance
		}
	}
	return 0
}

func deposit(t *testing.T, db *sqlx.DB, repo *repository.Repository, ledgerSvc *ledger.Service, userID, walletID uuid.UUID, currency models.Currency, amount int64) {
	t.Helper()
	ctx := context.Background()
	key := "dep-" + uuid.New().String()
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatalf("deposit begin tx: %v", err)
	}
	txn, err := repo.CreateTransaction(ctx, tx, userID, models.TransactionDeposit, models.StatusSuccessful, currency, amount, &key, map[string]interface{}{})
	if err != nil {
		tx.Rollback()
		t.Fatalf("deposit create txn: %v", err)
	}
	if err := ledgerSvc.RecordDeposit(ctx, tx, txn.ID, walletID, currency, amount); err != nil {
		tx.Rollback()
		t.Fatalf("deposit record: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("deposit commit: %v", err)
	}
}

func TestLedger_UnbalancedEntriesRejected(t *testing.T) {
	db := connectTestDB(t)
	_, walletID := seedUser(t, db)

	ctx := context.Background()
	repo := repository.New(db)
	ledgerSvc := ledger.NewService(repo)

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	wID := walletID
	err = ledgerSvc.WriteEntries(ctx, tx, uuid.New(), []ledger.Entry{
		{WalletID: &wID, Currency: models.USD, Amount: 5000, Direction: models.Credit, Account: "wallet:test"},
		{WalletID: nil, Currency: models.USD, Amount: 3000, Direction: models.Debit, Account: "external:test"},
	})

	if err == nil {
		t.Fatal("expected error for unbalanced entries, got nil")
	}
}

func TestDeposit_IdempotencyPreventsDoubleCredit(t *testing.T) {
	db := connectTestDB(t)
	userID, walletID := seedUser(t, db)

	ctx := context.Background()
	repo := repository.New(db)
	ledgerSvc := ledger.NewService(repo)

	key := "idem-" + uuid.New().String()

	tx1, _ := db.BeginTxx(ctx, nil)
	txn, err := repo.CreateTransaction(ctx, tx1, userID, models.TransactionDeposit, models.StatusSuccessful, models.USD, 10000, &key, map[string]interface{}{})
	if err != nil {
		tx1.Rollback()
		t.Fatalf("first deposit create txn: %v", err)
	}
	if err := ledgerSvc.RecordDeposit(ctx, tx1, txn.ID, walletID, models.USD, 10000); err != nil {
		tx1.Rollback()
		t.Fatalf("first deposit record: %v", err)
	}
	tx1.Commit()

	if bal := getBalance(t, repo, walletID, models.USD); bal != 10000 {
		t.Fatalf("after first deposit: want 10000, got %d", bal)
	}

	tx2, _ := db.BeginTxx(ctx, nil)
	_, err = repo.CreateTransaction(ctx, tx2, userID, models.TransactionDeposit, models.StatusSuccessful, models.USD, 10000, &key, map[string]interface{}{})
	tx2.Rollback()

	if err == nil {
		t.Fatal("expected unique constraint error on duplicate idempotency key, got nil")
	}

	if bal := getBalance(t, repo, walletID, models.USD); bal != 10000 {
		t.Errorf("after duplicate attempt: balance must not change, want 10000, got %d", bal)
	}
}

func TestPayout_InsufficientFundsRejected(t *testing.T) {
	db := connectTestDB(t)
	userID, walletID := seedUser(t, db)

	ctx := context.Background()
	repo := repository.New(db)
	ledgerSvc := ledger.NewService(repo)

	deposit(t, db, repo, ledgerSvc, userID, walletID, models.NGN, 500)

	payoutKey := "insuf-" + uuid.New().String()
	tx, _ := db.BeginTxx(ctx, nil)
	defer tx.Rollback()

	payoutTxn, err := repo.CreateTransaction(ctx, tx, userID, models.TransactionPayout, models.StatusPending, models.NGN, 1000, &payoutKey, map[string]interface{}{})
	if err != nil {
		t.Fatalf("create payout txn: %v", err)
	}

	err = ledgerSvc.RecordPayout(ctx, tx, payoutTxn.ID, walletID, models.NGN, 1000)
	if err == nil {
		t.Fatal("expected insufficient funds error, got nil")
	}
}
func TestPayout_FailureRestoresBalance(t *testing.T) {
	db := connectTestDB(t)
	userID, walletID := seedUser(t, db)

	ctx := context.Background()
	repo := repository.New(db)
	ledgerSvc := ledger.NewService(repo)

	const amount = int64(10000)
	deposit(t, db, repo, ledgerSvc, userID, walletID, models.NGN, amount)

	payoutKey := "payout-" + uuid.New().String()
	tx1, _ := db.BeginTxx(ctx, nil)
	payoutTxn, _ := repo.CreateTransaction(ctx, tx1, userID, models.TransactionPayout, models.StatusPending, models.NGN, amount, &payoutKey, map[string]interface{}{})
	if err := ledgerSvc.RecordPayout(ctx, tx1, payoutTxn.ID, walletID, models.NGN, amount); err != nil {
		tx1.Rollback()
		t.Fatalf("record payout: %v", err)
	}
	tx1.Commit()

	if bal := getBalance(t, repo, walletID, models.NGN); bal != 0 {
		t.Fatalf("after payout: want 0, got %d", bal)
	}

	reversalKey := "reversal-" + uuid.New().String()
	tx2, _ := db.BeginTxx(ctx, nil)
	reversalTxn, _ := repo.CreateTransaction(ctx, tx2, userID, models.TransactionPayout, models.StatusSuccessful, models.NGN, amount, &reversalKey, map[string]interface{}{})
	if err := ledgerSvc.RecordPayoutReversal(ctx, tx2, reversalTxn.ID, walletID, models.NGN, amount); err != nil {
		tx2.Rollback()
		t.Fatalf("record reversal: %v", err)
	}
	tx2.Commit()

	if bal := getBalance(t, repo, walletID, models.NGN); bal != amount {
		t.Errorf("after reversal: want %d, got %d", amount, bal)
	}
}

func TestConversion_BothBalancesUpdatedCorrectly(t *testing.T) {
	db := connectTestDB(t)
	userID, walletID := seedUser(t, db)

	ctx := context.Background()
	repo := repository.New(db)
	ledgerSvc := ledger.NewService(repo)

	deposit(t, db, repo, ledgerSvc, userID, walletID, models.USD, 10000)

	convKey := "conv-" + uuid.New().String()
	tx, _ := db.BeginTxx(ctx, nil)
	convTxn, _ := repo.CreateTransaction(ctx, tx, userID, models.TransactionConversion, models.StatusSuccessful, models.USD, 10000, &convKey, map[string]interface{}{})
	if err := ledgerSvc.RecordConversion(ctx, tx, convTxn.ID, walletID, models.USD, models.EUR, 10000, 8500); err != nil {
		tx.Rollback()
		t.Fatalf("record conversion: %v", err)
	}
	tx.Commit()

	if usd := getBalance(t, repo, walletID, models.USD); usd != 0 {
		t.Errorf("USD after conversion: want 0, got %d", usd)
	}
	if eur := getBalance(t, repo, walletID, models.EUR); eur != 8500 {
		t.Errorf("EUR after conversion: want 8500, got %d", eur)
	}
}
