package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"csp_report_backend/internal/db"

	"github.com/go-chi/chi/v5"
)

// ViolationHandler manages violation groups and triage actions.
type ViolationHandler struct {
	db *db.DB
}

// NewViolationHandler creates a new ViolationHandler.
func NewViolationHandler(database *db.DB) *ViolationHandler {
	return &ViolationHandler{db: database}
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

type BulkUpdateRequest struct {
	ViolationIDs []int64 `json:"violation_ids"`
	Status       string  `json:"status"`
}

var validStatuses = map[string]bool{
	"pending":           true,
	"approved_origin":   true,
	"approved_wildcard": true,
	"approved_self":     true,
	"rejected":          true,
	"ignored":           true,
}

// ListViolations handles GET /api/sessions/{id}/violations.
func (h *ViolationHandler) ListViolations(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Session ID is required", nil)
		return
	}

	directive := strings.TrimSpace(r.URL.Query().Get("directive"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	source := strings.TrimSpace(r.URL.Query().Get("source")) // "all", "self", "third_party"

	violations, err := h.db.ListViolationGroups(r.Context(), sessionID, directive, status, search, source)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve violations", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, violations)
}

// BulkApproveSelf handles POST /api/sessions/{id}/violations/approve-self.
func (h *ViolationHandler) BulkApproveSelf(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Session ID is required", nil)
		return
	}

	approvedCount, err := h.db.BulkApproveSelf(r.Context(), sessionID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to approve 'self' violations", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"approved": approvedCount,
		"updated":  approvedCount,
	})
}

// GetViolationSamples handles GET /api/sessions/{id}/violations/{violationId}/samples.
func (h *ViolationHandler) GetViolationSamples(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	rawID := chi.URLParam(r, "violationId")

	violationID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid violation ID", err.Error())
		return
	}

	vg, err := h.db.GetViolationGroup(r.Context(), violationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve violation group", err.Error())
		return
	}
	if vg == nil || vg.SessionID != sessionID {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Violation %d not found in session", violationID), nil)
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	samples, err := h.db.GetRawReportSamples(r.Context(), sessionID, vg.Directive, vg.OriginHost, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve report samples", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, samples)
}

// UpdateViolationStatus handles PATCH /api/sessions/{id}/violations/{violationId}.
func (h *ViolationHandler) UpdateViolationStatus(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	rawID := chi.URLParam(r, "violationId")

	violationID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid violation ID", err.Error())
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON body", err.Error())
		return
	}

	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	if !validStatuses[req.Status] {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", fmt.Sprintf("Invalid status '%s'", req.Status), nil)
		return
	}

	vg, err := h.db.GetViolationGroup(r.Context(), violationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check violation", err.Error())
		return
	}
	if vg == nil || vg.SessionID != sessionID {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Violation %d not found in session", violationID), nil)
		return
	}

	updated, err := h.db.UpdateViolationGroupStatus(r.Context(), violationID, req.Status)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update violation status", err.Error())
		return
	}
	if !updated {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Violation group not found", nil)
		return
	}

	updatedGroup, err := h.db.GetViolationGroup(r.Context(), violationID)
	if err != nil || updatedGroup == nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reload violation group", nil)
		return
	}

	WriteJSON(w, http.StatusOK, updatedGroup)
}

// BulkUpdateViolations handles POST /api/sessions/{id}/violations/bulk.
func (h *ViolationHandler) BulkUpdateViolations(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var req BulkUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON body", err.Error())
		return
	}

	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	if !validStatuses[req.Status] {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", fmt.Sprintf("Invalid status '%s'", req.Status), nil)
		return
	}

	if len(req.ViolationIDs) == 0 {
		WriteJSON(w, http.StatusOK, map[string]any{"updated": 0})
		return
	}

	updatedCount, err := h.db.BulkUpdateViolationGroupStatus(r.Context(), sessionID, req.ViolationIDs, req.Status)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to bulk update violations", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"updated": updatedCount,
	})
}
