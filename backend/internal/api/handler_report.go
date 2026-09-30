package api

import (
	"io"
	"net/http"

	"csp_report_backend/internal/db"
	"csp_report_backend/internal/parser"

	"github.com/go-chi/chi/v5"
)

// ReportHandler handles CSP report ingestion and CORS preflights.
type ReportHandler struct {
	db *db.DB
}

// NewReportHandler creates a new ReportHandler.
func NewReportHandler(database *db.DB) *ReportHandler {
	return &ReportHandler{db: database}
}

// HandleOptions responds to CORS preflight requests for report endpoints.
func (h *ReportHandler) HandleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, User-Agent, Authorization")
	w.WriteHeader(http.StatusNoContent)
}

// IngestReport processes incoming CSP reports for a session.
func (h *ReportHandler) IngestReport(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		sessionID = chi.URLParam(r, "id")
	}

	if sessionID == "" {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Session ID is required", nil)
		return
	}

	// Verify session exists
	session, err := h.db.GetSession(r.Context(), sessionID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check session", err.Error())
		return
	}
	if session == nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Session not found", nil)
		return
	}

	// Limit body to 5MB to prevent abuse
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Failed to read request body", err.Error())
		return
	}

	violations, err := parser.ParseReportPayload(body, session.TargetOrigin)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Failed to parse CSP report: "+err.Error(), nil)
		return
	}

	for _, v := range violations {
		if err := h.db.IngestReport(r.Context(), &v, session.ID); err != nil {
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to store report violation", err.Error())
			return
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusNoContent)
}
