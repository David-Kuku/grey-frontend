package ledger

import (
	"context"
	"fmt"

	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/models"
	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)
type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}
type Entry struct {
	WalletID    *uuid.UUID
	Currency    models.Currency
	Amount      int64
	Direction   models.LedgerDirection
	Account     string
	Description string
}
func (s *Service) WriteEntries(ctx context.Context, tx *sqlx.Tx, txnID uuid.UUID, entries []Entry) error {
	if len(entries) == 0 {
		return fmt.Errorf("ledger: at least one entry is required")
	}
	var checkSum int64
	for _, e := range entries {
		if e.Direction == models.Debit {
			checkSum += e.Amount
		} else {
			checkSum -= e.Amount
		}
	}
	if checkSum != 0 {
		return fmt.Errorf("ledger: entries do not balance (delta=%d)", checkSum)
	}

	for _, e := range entries {
		signed := e.Amount
		if e.Direction == models.Debit {
			signed = -e.Amount
		}

		entry := &models.LedgerEntry{
			ID:            uuid.New(),
			TransactionID: txnID,
			WalletID:      e.WalletID,
			Currency:      e.Currency,
			Amount:        e.Amount,
			Direction:     e.Direction,
			SignedAmount:  signed,
			Account:       e.Account,
			Description:   e.Description,
		}

		if err := s.repo.CreateLedgerEntry(ctx, tx, entry); err != nil {
			return fmt.Errorf("ledger: write entry: %w", err)
		}

		if e.WalletID != nil {
			bal, err := s.repo.GetBalanceForUpdate(ctx, tx, *e.WalletID, e.Currency)
			if err != nil {
				return fmt.Errorf("ledger: lock balance: %w", err)
			}

			newBalance := bal.Balance + signed
			if newBalance < 0 {
				return fmt.Errorf("ledger: insufficient funds in %s (have %d, need %d)", e.Currency, bal.Balance, e.Amount)
			}

			if err := s.repo.UpdateBalance(ctx, tx, *e.WalletID, e.Currency, newBalance); err != nil {
				return fmt.Errorf("ledger: update balance: %w", err)
			}
		}
	}

	return nil
}
func (s *Service) RecordDeposit(ctx context.Context, tx *sqlx.Tx, txnID uuid.UUID, walletID uuid.UUID, currency models.Currency, amount int64) error {
	entries := []Entry{
		{
			WalletID:    &walletID,
			Currency:    currency,
			Amount:      amount,
			Direction:   models.Credit,
			Account:     fmt.Sprintf("wallet:%s:%s", walletID, currency),
			Description: "Inbound deposit",
		},
		{
			WalletID:    nil,
			Currency:    currency,
			Amount:      amount,
			Direction:   models.Debit,
			Account:     "external:banking_partner",
			Description: "Inbound deposit from banking partner",
		},
	}
	return s.WriteEntries(ctx, tx, txnID, entries)
}
func (s *Service) RecordConversion(ctx context.Context, tx *sqlx.Tx, txnID uuid.UUID, walletID uuid.UUID, sourceCurrency, targetCurrency models.Currency, sourceAmount, targetAmount int64) error {
	leg1 := []Entry{
		{
			WalletID:    &walletID,
			Currency:    sourceCurrency,
			Amount:      sourceAmount,
			Direction:   models.Debit,
			Account:     fmt.Sprintf("wallet:%s:%s", walletID, sourceCurrency),
			Description: fmt.Sprintf("FX conversion out: %s → %s", sourceCurrency, targetCurrency),
		},
		{
			WalletID:    nil,
			Currency:    sourceCurrency,
			Amount:      sourceAmount,
			Direction:   models.Credit,
			Account:     "house:fx_pool",
			Description: fmt.Sprintf("FX conversion: received %s from user", sourceCurrency),
		},
	}
	if err := s.WriteEntries(ctx, tx, txnID, leg1); err != nil {
		return err
	}
	leg2 := []Entry{
		{
			WalletID:    nil,
			Currency:    targetCurrency,
			Amount:      targetAmount,
			Direction:   models.Debit,
			Account:     "house:fx_pool",
			Description: fmt.Sprintf("FX conversion: sent %s to user", targetCurrency),
		},
		{
			WalletID:    &walletID,
			Currency:    targetCurrency,
			Amount:      targetAmount,
			Direction:   models.Credit,
			Account:     fmt.Sprintf("wallet:%s:%s", walletID, targetCurrency),
			Description: fmt.Sprintf("FX conversion in: %s → %s", sourceCurrency, targetCurrency),
		},
	}
	return s.WriteEntries(ctx, tx, txnID, leg2)
}
func (s *Service) RecordPayout(ctx context.Context, tx *sqlx.Tx, txnID uuid.UUID, walletID uuid.UUID, currency models.Currency, amount int64) error {
	entries := []Entry{
		{
			WalletID:    &walletID,
			Currency:    currency,
			Amount:      amount,
			Direction:   models.Debit,
			Account:     fmt.Sprintf("wallet:%s:%s", walletID, currency),
			Description: "Outbound payout",
		},
		{
			WalletID:    nil,
			Currency:    currency,
			Amount:      amount,
			Direction:   models.Credit,
			Account:     "external:payout_rail",
			Description: "Outbound payout to recipient bank",
		},
	}
	return s.WriteEntries(ctx, tx, txnID, entries)
}
func (s *Service) RecordPayoutReversal(ctx context.Context, tx *sqlx.Tx, txnID uuid.UUID, walletID uuid.UUID, currency models.Currency, amount int64) error {
	entries := []Entry{
		{
			WalletID:    &walletID,
			Currency:    currency,
			Amount:      amount,
			Direction:   models.Credit,
			Account:     fmt.Sprintf("wallet:%s:%s", walletID, currency),
			Description: "Payout reversal: funds returned",
		},
		{
			WalletID:    nil,
			Currency:    currency,
			Amount:      amount,
			Direction:   models.Debit,
			Account:     "external:payout_rail",
			Description: "Payout reversal: failed payout returned",
		},
	}
	return s.WriteEntries(ctx, tx, txnID, entries)
}
func (s *Service) VerifyBalance(ctx context.Context, walletID uuid.UUID, currency models.Currency) error {
	balances, err := s.repo.GetBalances(ctx, walletID)
	if err != nil {
		return err
	}

	for _, bal := range balances {
		if bal.Currency == currency {
			ledgerBal, err := s.repo.GetLedgerBalance(ctx, walletID, currency)
			if err != nil {
				return err
			}
			if bal.Balance != ledgerBal {
				return fmt.Errorf("balance mismatch for %s: cached=%d, ledger=%d", currency, bal.Balance, ledgerBal)
			}
			return nil
		}
	}
	return fmt.Errorf("no balance found for currency %s", currency)
}
