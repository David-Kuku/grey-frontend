package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/fx"
	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/ledger"
	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/middleware"
	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/models"
	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/repository"
	"github.com/google/uuid"
)

type ConversionHandler struct {
	repo          *repository.Repository
	fxService     *fx.Service
	ledgerService *ledger.Service
	logger        *slog.Logger
}

func NewConversionHandler(repo *repository.Repository, fxService *fx.Service, ledgerService *ledger.Service, logger *slog.Logger) *ConversionHandler {
	return &ConversionHandler{repo: repo, fxService: fxService, ledgerService: ledgerService, logger: logger}
}
func (h *ConversionHandler) GetQuote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req models.QuoteRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON")
		return
	}
	errors := make(map[string]string)
	if !req.SourceCurrency.IsValid() {
		errors["source_currency"] = "Source currency must be one of: USD, GBP, EUR, NGN, KES"
	}
	if !req.TargetCurrency.IsValid() {
		errors["target_currency"] = "Target currency must be one of: USD, GBP, EUR, NGN, KES"
	}
	if req.SourceCurrency == req.TargetCurrency {
		errors["target_currency"] = "Target currency must differ from source currency"
	}
	if req.SourceAmount <= 0 {
		errors["source_amount"] = "Source amount must be a positive integer (in minor units)"
	}
	if len(errors) > 0 {
		respondErrorWithDetails(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid quote request", errors)
		return
	}
	wallet, err := h.repo.GetWalletByUserID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve wallet")
		return
	}

	balances, err := h.repo.GetBalances(r.Context(), wallet.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve balances")
		return
	}

	var sourceBalance int64
	for _, b := range balances {
		if b.Currency == req.SourceCurrency {
			sourceBalance = b.Balance
			break
		}
	}

	if sourceBalance < req.SourceAmount {
		respondError(w, http.StatusUnprocessableEntity, "INSUFFICIENT_BALANCE",
			"Insufficient balance in source currency. Available: "+formatMinorUnits(sourceBalance, req.SourceCurrency))
		return
	}
	quote, err := h.fxService.CreateQuote(r.Context(), userID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "FX_ERROR", "Failed to generate FX quote: "+err.Error())
		return
	}

	expiresIn := int(time.Until(quote.ExpiresAt).Seconds())

	h.logger.Info("fx quote created", "quote_id", quote.ID, "user_id", userID, "source_currency", req.SourceCurrency, "target_currency", req.TargetCurrency, "source_amount", req.SourceAmount, "request_id", middleware.GetRequestID(r.Context()))
	respondJSON(w, http.StatusOK, models.QuoteResponse{
		QuoteID:        quote.ID.String(),
		SourceCurrency: quote.SourceCurrency,
		TargetCurrency: quote.TargetCurrency,
		SourceAmount:   quote.SourceAmount,
		TargetAmount:   quote.TargetAmount,
		MarketRate:     quote.MarketRate,
		QuotedRate:     quote.QuotedRate,
		SpreadPct:      quote.SpreadPct,
		FeeAmount:      quote.FeeAmount,
		FeeCurrency:    *quote.FeeCurrency,
		ExpiresAt:      quote.ExpiresAt.Format(time.RFC3339),
		ExpiresInSecs:  expiresIn,
	})
}
func (h *ConversionHandler) ExecuteConversion(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req models.ExecuteConversionRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON")
		return
	}

	if req.QuoteID == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "quote_id is required")
		return
	}

	quoteUUID, err := uuid.Parse(req.QuoteID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "quote_id must be a valid UUID")
		return
	}
	quote, err := h.fxService.GetQuote(r.Context(), quoteUUID, userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "QUOTE_NOT_FOUND", "Quote not found or does not belong to you")
		return
	}

	if quote.Executed {
		respondError(w, http.StatusConflict, "QUOTE_ALREADY_EXECUTED", "This quote has already been executed")
		return
	}

	if quote.IsExpired() {
		respondError(w, http.StatusGone, "QUOTE_EXPIRED", "This quote has expired. Please request a new one.")
		return
	}
	wallet, err := h.repo.GetWalletByUserID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve wallet")
		return
	}
	tx, err := h.repo.DB().BeginTxx(r.Context(), nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process conversion")
		return
	}
	defer tx.Rollback()
	sourceBal, err := h.repo.GetBalanceForUpdate(r.Context(), tx, wallet.ID, quote.SourceCurrency)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to lock balance")
		return
	}

	if sourceBal.Balance < quote.SourceAmount {
		respondError(w, http.StatusUnprocessableEntity, "INSUFFICIENT_BALANCE",
			"Insufficient balance. Your balance may have changed since the quote was issued.")
		return
	}
	metadata := map[string]interface{}{
		"quote_id":        quote.ID,
		"target_currency": quote.TargetCurrency,
		"target_amount":   quote.TargetAmount,
		"market_rate":     quote.MarketRate,
		"quoted_rate":     quote.QuotedRate,
	}

	txn, err := h.repo.CreateTransaction(r.Context(), tx, userID,
		models.TransactionConversion, models.StatusSuccessful, quote.SourceCurrency, quote.SourceAmount, nil, metadata)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create conversion record")
		return
	}
	if err := h.ledgerService.RecordConversion(r.Context(), tx, txn.ID, wallet.ID,
		quote.SourceCurrency, quote.TargetCurrency, quote.SourceAmount, quote.TargetAmount); err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record conversion: "+err.Error())
		return
	}
	if err := h.repo.MarkQuoteExecuted(r.Context(), tx, quote.ID, txn.ID); err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark quote as executed")
		return
	}

	if err := tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to finalize conversion")
		return
	}
	balances, _ := h.repo.GetBalances(r.Context(), wallet.ID)
	var sourceBalance, targetBalance int64
	for _, b := range balances {
		if b.Currency == quote.SourceCurrency {
			sourceBalance = b.Balance
		}
		if b.Currency == quote.TargetCurrency {
			targetBalance = b.Balance
		}
	}

	h.logger.Info("conversion executed", "transaction_id", txn.ID, "user_id", userID, "quote_id", quote.ID, "source_currency", quote.SourceCurrency, "target_currency", quote.TargetCurrency, "source_amount", quote.SourceAmount, "target_amount", quote.TargetAmount, "request_id", middleware.GetRequestID(r.Context()))
	respondJSON(w, http.StatusOK, models.ExecuteConversionResponse{
		Transaction:    *txn,
		SourceBalance:  sourceBalance,
		TargetBalance:  targetBalance,
		SourceCurrency: quote.SourceCurrency,
		TargetCurrency: quote.TargetCurrency,
	})
}
func formatMinorUnits(amount int64, currency models.Currency) string {
	major := amount / currency.MinorUnits()
	minor := amount % currency.MinorUnits()
	return string(currency) + " " + intToStr(major) + "." + padLeft(intToStr(minor), 2, '0')
}

func intToStr(n int64) string {
	if n < 0 {
		n = -n
	}
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func padLeft(s string, length int, pad byte) string {
	for len(s) < length {
		s = string(pad) + s
	}
	return s
}
