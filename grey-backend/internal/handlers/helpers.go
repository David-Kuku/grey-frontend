package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/models"
)
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, models.APIError{
		Code:    code,
		Message: message,
	})
}
func respondErrorWithDetails(w http.ResponseWriter, status int, code, message string, details map[string]string) {
	respondJSON(w, status, models.APIError{
		Code:    code,
		Message: message,
		Details: details,
	})
}
func decodeJSON(r *http.Request, target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}
