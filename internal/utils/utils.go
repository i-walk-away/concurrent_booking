package utils

import (
	"encoding/json"
	"net/http"
)

// WriteJSON writes v as a JSON response with the given HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// The response headers have already been written, so there is no useful
	// HTTP error response we can send if JSON encoding fails.
	_ = json.NewEncoder(w).Encode(v)
}
