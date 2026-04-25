package handlers

import (
	"database/sql"
	"net/http"
	"regexp"
	"strings"

	"github.com/David-Kuku/grey-frontend/grey-backend/internal/auth"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/models"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/repository"
)

type AuthHandler struct {
	repo        *repository.Repository
	authService *auth.Service
}

func NewAuthHandler(repo *repository.Repository, authService *auth.Service) *AuthHandler {
	return &AuthHandler{repo: repo, authService: authService}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req models.SignupRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON")
		return
	}
	errors := make(map[string]string)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" {
		errors["email"] = "Email is required"
	} else if !emailRegex.MatchString(req.Email) {
		errors["email"] = "Email format is invalid"
	}
	if len(req.Password) < 8 {
		errors["password"] = "Password must be at least 8 characters"
	}
	if len(errors) > 0 {
		respondErrorWithDetails(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", errors)
		return
	}
	existing, err := h.repo.GetUserByEmail(r.Context(), req.Email)
	if err == nil && existing != nil {
		respondError(w, http.StatusConflict, "EMAIL_EXISTS", "An account with this email already exists")
		return
	}
	hash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process registration")
		return
	}
	tx, err := h.repo.DB().BeginTxx(r.Context(), nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process registration")
		return
	}
	defer tx.Rollback()

	user, err := h.repo.CreateUser(r.Context(), tx, req.Email, hash)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			respondError(w, http.StatusConflict, "EMAIL_EXISTS", "An account with this email already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create account")
		return
	}
	_, err = h.repo.CreateWallet(r.Context(), tx, user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create wallet")
		return
	}

	if err := tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to finalize registration")
		return
	}

	respondJSON(w, http.StatusCreated, user)
}
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email and password are required")
		return
	}

	user, err := h.repo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "no rows") {
			respondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process login")
		return
	}

	if err := h.authService.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		respondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}

	token, err := h.authService.GenerateToken(user.ID, user.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate authentication token")
		return
	}

	respondJSON(w, http.StatusOK, models.AuthResponse{
		Token: token,
		User:  *user,
	})
}
