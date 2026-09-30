package api

import (
	"encoding/json"
	"net/http"

	"csp_report_backend/internal/model"

	"github.com/go-chi/cors"
)

// WriteJSON sends a JSON response with status code.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// WriteError sends a standard REST error envelope.
func WriteError(w http.ResponseWriter, status int, code, message string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := model.ErrorResponse{
		Error: model.ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// CorsMiddleware returns a configured CORS middleware handler.
func CorsMiddleware() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "User-Agent"},
		ExposedHeaders:   []string{"Link", "Location"},
		AllowCredentials: false,
		MaxAge:           300,
	})
}
