package handlers

import (
	"net/http"
	"strings"

	"github.com/David-Kuku/grey-frontend/grey-backend/internal/ledger"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/middleware"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/models"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/repository"
)

type DepositHandler struct {
	repo          *repository.Repository
	ledgerService *ledger.Service
}

func NewDepositHandler(repo *repository.Repository, ledgerService *ledger.Service) *DepositHandler {
	return &DepositHandler{repo: repo, ledgerService: ledgerService}
}
func (h *DepositHandler) CreateDeposit(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))

	var req models.DepositRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON")
		return
	}
	errors := make(map[string]string)
	if !req.Currency.IsValid() {
		errors["currency"] = "Currency must be one of: USD, GBP, EUR, NGN, KES"
	}
	if req.Amount <= 0 {
		errors["amount"] = "Amount must be a positive integer (in minor units, e.g. cents)"
	}
	if idempotencyKey == "" {
		errors["Idempotency-Key"] = "Idempotency-Key header is required to prevent duplicate deposits"
	}
	if len(errors) > 0 {
		respondErrorWithDetails(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid deposit requestaas", errors)
		return
	}
	existing, err := h.repo.GetTransactionByIdempotencyKey(r.Context(), idempotencyKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check idempotency")
		return
	}
	if existing != nil {
		respondJSON(w, http.StatusOK, models.DepositResponse{
			Transaction: *existing,
		})
		return
	}
	wallet, err := h.repo.GetWalletByUserID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve wallet")
		return
	}
	tx, err := h.repo.DB().BeginTxx(r.Context(), nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process deposit")
		return
	}
	defer tx.Rollback()
	metadata := map[string]interface{}{}
	txn, err := h.repo.CreateTransaction(r.Context(), tx, userID,
		models.TransactionDeposit, models.StatusSuccessful, req.Currency, req.Amount, &idempotencyKey, metadata)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create deposit record")
		return
	}
	if err := h.ledgerService.RecordDeposit(r.Context(), tx, txn.ID, wallet.ID, req.Currency, req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record deposit in ledger")
		return
	}

	if err := tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to finalize deposit")
		return
	}

	respondJSON(w, http.StatusCreated, models.DepositResponse{
		Transaction: *txn,
	})
}
