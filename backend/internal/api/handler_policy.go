package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"csp_report_backend/internal/db"
	"csp_report_backend/internal/generator"
	"csp_report_backend/internal/model"

	"github.com/go-chi/chi/v5"
)

// PolicyHandler handles policy generation, export, and settings.
type PolicyHandler struct {
	db *db.DB
}

// NewPolicyHandler creates a new PolicyHandler.
func NewPolicyHandler(database *db.DB) *PolicyHandler {
	return &PolicyHandler{db: database}
}

// GetPolicy handles GET /api/sessions/{id}/policy (and /export).
func (h *PolicyHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Session ID is required", nil)
		return
	}

	session, err := h.db.GetSession(r.Context(), sessionID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve session", err.Error())
		return
	}
	if session == nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Session '%s' not found", sessionID), nil)
		return
	}

	settings, err := h.db.GetSessionPolicySetting(r.Context(), sessionID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve policy settings", err.Error())
		return
	}
	if settings == nil {
		settings = &model.SessionPolicySetting{
			SessionID:      sessionID,
			DefaultSrc:     "'none'",
			FormAction:     "'none'",
			FrameAncestors: "'none'",
		}
	}

	violations, err := h.db.ListViolationGroups(r.Context(), sessionID, "", "", "", "")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve violations", err.Error())
		return
	}

	var approved []model.ViolationGroup
	for _, v := range violations {
		if strings.HasPrefix(v.Status, "approved") {
			approved = append(approved, v)
		}
	}

	policyOut := generator.GeneratePolicy(settings, approved)

	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	switch format {
	case "raw":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(policyOut.Raw))
	case "header":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(policyOut.Header))
	case "meta":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(policyOut.Meta))
	case "nginx":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(policyOut.Nginx))
	case "apache":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(policyOut.Apache))
	default:
		WriteJSON(w, http.StatusOK, policyOut)
	}
}

// UpdateSettings handles PUT /api/sessions/{id}/settings.
func (h *PolicyHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Session ID is required", nil)
		return
	}

	session, err := h.db.GetSession(r.Context(), sessionID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve session", err.Error())
		return
	}
	if session == nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Session '%s' not found", sessionID), nil)
		return
	}

	var req model.SessionPolicySetting
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON body", err.Error())
		return
	}
	req.SessionID = sessionID

	if err := h.db.UpdateSessionPolicySetting(r.Context(), &req); err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save policy settings", err.Error())
		return
	}

	updated, err := h.db.GetSessionPolicySetting(r.Context(), sessionID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reload policy settings", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, updated)
}
