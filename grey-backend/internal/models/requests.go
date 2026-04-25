package models

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type DepositRequest struct {
	Currency Currency `json:"currency"`
	Amount   int64    `json:"amount"`
}

type DepositResponse struct {
	Transaction Transaction `json:"transaction"`
}

type QuoteRequest struct {
	SourceCurrency Currency `json:"source_currency"`
	TargetCurrency Currency `json:"target_currency"`
	SourceAmount   int64    `json:"source_amount"`
}

type QuoteResponse struct {
	QuoteID        string   `json:"quote_id"`
	SourceCurrency Currency `json:"source_currency"`
	TargetCurrency Currency `json:"target_currency"`
	SourceAmount   int64    `json:"source_amount"`
	TargetAmount   int64    `json:"target_amount"`
	MarketRate     string   `json:"market_rate"`
	QuotedRate     string   `json:"quoted_rate"`
	SpreadPct      string   `json:"spread_pct"`
	FeeAmount      int64    `json:"fee_amount"`
	FeeCurrency    Currency `json:"fee_currency"`
	ExpiresAt      string   `json:"expires_at"`
	ExpiresInSecs  int      `json:"expires_in_seconds"`
}

type ExecuteConversionRequest struct {
	QuoteID string `json:"quote_id"`
}

type ExecuteConversionResponse struct {
	Transaction    Transaction `json:"transaction"`
	SourceBalance  int64       `json:"source_balance"`
	TargetBalance  int64       `json:"target_balance"`
	SourceCurrency Currency    `json:"source_currency"`
	TargetCurrency Currency    `json:"target_currency"`
}

type PayoutRequest struct {
	SourceCurrency    Currency `json:"source_currency"`
	Amount            int64    `json:"amount"`
	RecipientName     string   `json:"recipient_name"`
	RecipientBankCode string   `json:"recipient_bank_code"`
	RecipientAccount  string   `json:"recipient_account"`
}

type PayoutResponse struct {
	Payout      Payout      `json:"payout"`
	Transaction Transaction `json:"transaction"`
	NewBalance  int64       `json:"new_balance"`
}

type TransactionHistoryParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type TransactionSummary struct {
	Transaction
	AmountFormatted string `json:"amount_formatted"`
}

type TransactionHistoryResponse struct {
	Transactions []TransactionSummary `json:"transactions"`
	Page         int                  `json:"page"`
	PageSize     int                  `json:"page_size"`
	TotalCount   int                  `json:"total_count"`
	HasMore      bool                 `json:"has_more"`
}

type BalancesResponse struct {
	Balances []BalanceResponse `json:"balances"`
}

type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}
