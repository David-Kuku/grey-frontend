package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/David-Kuku/grey-frontend/grey-backend/internal/ledger"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/middleware"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/models"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/payout"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/repository"
	"github.com/google/uuid"
)

type PayoutHandler struct {
	repo          *repository.Repository
	ledgerService *ledger.Service
	payoutService *payout.Service
	logger        *slog.Logger
}

func NewPayoutHandler(repo *repository.Repository, ledgerService *ledger.Service, payoutService *payout.Service, logger *slog.Logger) *PayoutHandler {
	return &PayoutHandler{repo: repo, ledgerService: ledgerService, payoutService: payoutService, logger: logger}
}
func (h *PayoutHandler) CreatePayout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))

	var req models.PayoutRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON")
		return
	}
	errors := make(map[string]string)
	if !req.SourceCurrency.IsPayoutCurrency() {
		errors["source_currency"] = "Payouts are only supported in NGN and KES"
	}
	if req.Amount <= 0 {
		errors["amount"] = "Amount must be a positive integer (in minor units)"
	}
	if strings.TrimSpace(req.RecipientName) == "" {
		errors["recipient_name"] = "Recipient name is required"
	}
	if strings.TrimSpace(req.RecipientBankCode) == "" {
		errors["recipient_bank_code"] = "Recipient bank code is required"
	}
	if strings.TrimSpace(req.RecipientAccount) == "" {
		errors["recipient_account"] = "Recipient account number is required"
	}
	if idempotencyKey == "" {
		errors["Idempotency-Key"] = "Idempotency-Key header is required"
	}
	if len(errors) > 0 {
		respondErrorWithDetails(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid payout request", errors)
		return
	}
	existing, err := h.repo.GetTransactionByIdempotencyKey(r.Context(), idempotencyKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check idempotency")
		return
	}
	if existing != nil {
		p, _ := h.repo.GetPayoutByTransactionID(r.Context(), existing.ID)
		if p != nil {
			respondJSON(w, http.StatusOK, models.PayoutResponse{
				Payout:      *p,
				Transaction: *existing,
			})
			return
		}
	}
	wallet, err := h.repo.GetWalletByUserID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve wallet")
		return
	}
	tx, err := h.repo.DB().BeginTxx(r.Context(), nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process payout")
		return
	}
	defer tx.Rollback()
	bal, err := h.repo.GetBalanceForUpdate(r.Context(), tx, wallet.ID, req.SourceCurrency)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to lock balance")
		return
	}

	if bal.Balance < req.Amount {
		respondError(w, http.StatusUnprocessableEntity, "INSUFFICIENT_BALANCE",
			"Insufficient balance in "+string(req.SourceCurrency))
		return
	}
	metadata := map[string]interface{}{
		"recipient_name":    req.RecipientName,
		"recipient_bank":    req.RecipientBankCode,
		"recipient_account": req.RecipientAccount,
	}

	txn, err := h.repo.CreateTransaction(r.Context(), tx, userID,
		models.TransactionPayout, models.StatusPending, req.SourceCurrency, req.Amount, &idempotencyKey, metadata)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create payout record")
		return
	}
	if err := h.ledgerService.RecordPayout(r.Context(), tx, txn.ID, wallet.ID, req.SourceCurrency, req.Amount); err != nil {
		if strings.Contains(err.Error(), "insufficient balance") {
			respondError(w, http.StatusUnprocessableEntity, "INSUFFICIENT_BALANCE", err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record payout in ledger")
		return
	}
	p := &models.Payout{
		ID:                uuid.New(),
		TransactionID:     txn.ID,
		UserID:            userID,
		SourceCurrency:    req.SourceCurrency,
		Amount:            req.Amount,
		RecipientName:     req.RecipientName,
		RecipientBankCode: req.RecipientBankCode,
		RecipientAccount:  req.RecipientAccount,
		Status:            models.PayoutPending,
	}

	if err := h.repo.CreatePayout(r.Context(), tx, p); err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create payout")
		return
	}

	if err := tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to finalize payout")
		return
	}
	balances, _ := h.repo.GetBalances(r.Context(), wallet.ID)
	var newBalance int64
	for _, b := range balances {
		if b.Currency == req.SourceCurrency {
			newBalance = b.Balance
			break
		}
	}
	if err := h.payoutService.EnqueuePayout(r.Context(), p.ID); err != nil {
		h.logger.Error("failed to enqueue payout job", "error", err, "payout_id", p.ID)
	}

	respondJSON(w, http.StatusCreated, models.PayoutResponse{
		Payout:      *p,
		Transaction: *txn,
		NewBalance:  newBalance,
	})
}