package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/David-Kuku/grey-frontend/grey-backend/internal/middleware"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/models"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/repository"
)

type WalletHandler struct {
	repo *repository.Repository
}

func NewWalletHandler(repo *repository.Repository) *WalletHandler {
	return &WalletHandler{repo: repo}
}
func (h *WalletHandler) GetBalances(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

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

	response := make([]models.BalanceResponse, 0, len(balances))
	for _, b := range balances {
		major := b.Balance / b.Currency.MinorUnits()
		minor := b.Balance % b.Currency.MinorUnits()
		response = append(response, models.BalanceResponse{
			Currency:         b.Currency,
			BalanceMinor:     b.Balance,
			BalanceFormatted: fmt.Sprintf("%d.%02d", major, minor),
		})
	}

	respondJSON(w, http.StatusOK, models.BalancesResponse{Balances: response})
}
func (h *WalletHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	errors := make(map[string]string)

	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			errors["page"] = "page must be a positive integer"
		} else {
			page = v
		}
	}

	pageSize := 20
	if raw := r.URL.Query().Get("page_size"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			errors["page_size"] = "page_size must be a positive integer"
		} else if v > 100 {
			errors["page_size"] = "page_size cannot exceed 100"
		} else {
			pageSize = v
		}
	}

	if len(errors) > 0 {
		respondErrorWithDetails(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid pagination parameters", errors)
		return
	}

	txns, total, err := h.repo.GetTransactionHistory(r.Context(), userID, page, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve transactions")
		return
	}

	summaries := make([]models.TransactionSummary, len(txns))
	for i, t := range txns {
		major := t.Amount / t.Currency.MinorUnits()
		minor := t.Amount % t.Currency.MinorUnits()
		summaries[i] = models.TransactionSummary{
			Transaction:     t,
			AmountFormatted: fmt.Sprintf("%s %d.%02d", t.Currency, major, minor),
		}
	}

	hasMore := (page * pageSize) < total

	respondJSON(w, http.StatusOK, models.TransactionHistoryResponse{
		Transactions: summaries,
		Page:         page,
		PageSize:     pageSize,
		TotalCount:   total,
		HasMore:      hasMore,
	})
}

func (h *WalletHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	txID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_TRANSACTION_ID", "Transaction ID must be a valid UUID")
		return
	}

	txn, err := h.repo.GetTransactionByID(r.Context(), userID, txID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve transaction")
		return
	}
	if txn == nil {
		respondError(w, http.StatusNotFound, "TRANSACTION_NOT_FOUND", "Transaction not found")
		return
	}

	major := txn.Amount / txn.Currency.MinorUnits()
	minor := txn.Amount % txn.Currency.MinorUnits()
	respondJSON(w, http.StatusOK, models.TransactionSummary{
		Transaction:     *txn,
		AmountFormatted: fmt.Sprintf("%s %d.%02d", txn.Currency, major, minor),
	})
}
