package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Currency string

const (
	USD Currency = "USD"
	GBP Currency = "GBP"
	EUR Currency = "EUR"
	NGN Currency = "NGN"
	KES Currency = "KES"
)

var SupportedCurrencies = []Currency{USD, GBP, EUR, NGN, KES}
func (c Currency) MinorUnits() int64 {
	return 100
}

func (c Currency) IsValid() bool {
	for _, sc := range SupportedCurrencies {
		if c == sc {
			return true
		}
	}
	return false
}
func (c Currency) IsPayoutCurrency() bool {
	return c == NGN || c == KES
}

type User struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type Wallet struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type WalletBalance struct {
	ID        uuid.UUID `db:"id" json:"id"`
	WalletID  uuid.UUID `db:"wallet_id" json:"wallet_id"`
	Currency  Currency  `db:"currency" json:"currency"`
	Balance   int64     `db:"balance" json:"balance"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
type BalanceResponse struct {
	Currency         Currency `json:"currency"`
	BalanceMinor     int64    `json:"balance_minor"`
	BalanceFormatted string   `json:"balance"`
}

type TransactionType string

const (
	TransactionDeposit    TransactionType = "deposit"
	TransactionConversion TransactionType = "conversion"
	TransactionPayout     TransactionType = "payout"
)

type TransactionStatus string

const (
	StatusPending    TransactionStatus = "pending"
	StatusProcessing TransactionStatus = "processing"
	StatusSuccessful TransactionStatus = "successful"
	StatusFailed     TransactionStatus = "failed"
	StatusReversed   TransactionStatus = "reversed"
)

type Transaction struct {
	ID              uuid.UUID         `db:"id" json:"id"`
	UserID          uuid.UUID         `db:"user_id" json:"user_id"`
	TransactionType TransactionType   `db:"type" json:"transaction_type"`
	Status          TransactionStatus `db:"status" json:"status"`
	Currency        Currency          `db:"currency" json:"currency"`
	Amount          int64             `db:"amount" json:"amount"`
	IdempotencyKey  *string           `db:"idempotency_key" json:"idempotency_key,omitempty"`
	Metadata        json.RawMessage   `db:"metadata" json:"metadata"`
	CreatedAt       time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time         `db:"updated_at" json:"updated_at"`
}

type LedgerDirection string

const (
	Debit  LedgerDirection = "debit"
	Credit LedgerDirection = "credit"
)

type LedgerEntry struct {
	ID            uuid.UUID       `db:"id" json:"id"`
	TransactionID uuid.UUID       `db:"transaction_id" json:"transaction_id"`
	WalletID      *uuid.UUID      `db:"wallet_id" json:"wallet_id,omitempty"`
	Currency      Currency        `db:"currency" json:"currency"`
	Amount        int64           `db:"amount" json:"amount"`
	Direction     LedgerDirection `db:"direction" json:"direction"`
	SignedAmount  int64           `db:"signed_amount" json:"signed_amount"`
	Account       string          `db:"account" json:"account"`
	Description   string          `db:"description" json:"description"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
}

type FXQuote struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	UserID         uuid.UUID  `db:"user_id" json:"user_id"`
	SourceCurrency Currency   `db:"source_currency" json:"source_currency"`
	TargetCurrency Currency   `db:"target_currency" json:"target_currency"`
	MarketRate     string     `db:"market_rate" json:"market_rate"`
	QuotedRate     string     `db:"quoted_rate" json:"quoted_rate"`
	SpreadPct      string     `db:"spread_pct" json:"spread_pct"`
	SourceAmount   int64      `db:"source_amount" json:"source_amount"`
	TargetAmount   int64      `db:"target_amount" json:"target_amount"`
	FeeAmount      int64      `db:"fee_amount" json:"fee_amount"`
	FeeCurrency    *Currency  `db:"fee_currency" json:"fee_currency,omitempty"`
	ExpiresAt      time.Time  `db:"expires_at" json:"expires_at"`
	Executed       bool       `db:"executed" json:"executed"`
	ExecutedAt     *time.Time `db:"executed_at" json:"executed_at,omitempty"`
	TransactionID  *uuid.UUID `db:"transaction_id" json:"transaction_id,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
}

func (q *FXQuote) IsExpired() bool {
	return time.Now().After(q.ExpiresAt)
}

type PayoutStatus string

const (
	PayoutPending    PayoutStatus = "pending"
	PayoutProcessing PayoutStatus = "processing"
	PayoutSuccessful PayoutStatus = "successful"
	PayoutFailed     PayoutStatus = "failed"
)

type Payout struct {
	ID                    uuid.UUID    `db:"id" json:"id"`
	TransactionID         uuid.UUID    `db:"transaction_id" json:"transaction_id"`
	UserID                uuid.UUID    `db:"user_id" json:"user_id"`
	SourceCurrency        Currency     `db:"source_currency" json:"source_currency"`
	Amount                int64        `db:"amount" json:"amount"`
	RecipientName         string       `db:"recipient_name" json:"recipient_name"`
	RecipientBankCode     string       `db:"recipient_bank_code" json:"recipient_bank_code"`
	RecipientAccount      string       `db:"recipient_account" json:"recipient_account"`
	Status                PayoutStatus `db:"status" json:"status"`
	FailureReason         *string      `db:"failure_reason" json:"failure_reason,omitempty"`
	ReversalTransactionID *uuid.UUID   `db:"reversal_transaction_id" json:"reversal_transaction_id,omitempty"`
	CreatedAt             time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time    `db:"updated_at" json:"updated_at"`
}

type TransactionHistoryEntry struct {
	ID              uuid.UUID         `json:"id"`
	Type            TransactionType   `json:"type"`
	Status          TransactionStatus `json:"status"`
	Currency        Currency          `json:"currency"`
	Amount          int64             `json:"amount"`
	AmountFormatted string            `json:"amount_formatted"`
	TargetCurrency *Currency `json:"target_currency,omitempty"`
	TargetAmount   *int64    `json:"target_amount,omitempty"`
	Rate           *string   `json:"rate,omitempty"`
	RecipientName *string `json:"recipient_name,omitempty"`
	RecipientBank *string `json:"recipient_bank,omitempty"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
