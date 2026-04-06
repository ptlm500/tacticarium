package handler

import (
	"encoding/json"
	"net/http"
)

// writeJSON is used by raw chi handlers (OAuth, imports) that haven't been converted to huma.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
