package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"csp_report_backend/internal/db"
	"csp_report_backend/internal/model"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// SessionHandler handles session management endpoints.
type SessionHandler struct {
	db *db.DB
}

// NewSessionHandler creates a new SessionHandler.
func NewSessionHandler(database *db.DB) *SessionHandler {
	return &SessionHandler{db: database}
}

type CreateSessionRequest struct {
	Name         string `json:"name"`
	TargetOrigin string `json:"target_origin"`
	Description  string `json:"description"`
}

type SessionDetailResponse struct {
	model.Session
	PolicySetting *model.SessionPolicySetting `json:"policy_setting,omitempty"`
}

// ListSessions handles GET /api/sessions.
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.db.ListSessions(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve sessions", err.Error())
		return
	}

	scheme, host := getRequestHost(r)
	for i := range sessions {
		enrichSessionMetadata(&sessions[i], scheme, host)
	}

	WriteJSON(w, http.StatusOK, sessions)
}

// CreateSession handles POST /api/sessions.
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body JSON", err.Error())
		return
	}

	name := strings.TrimSpace(req.Name)
	targetOrigin := strings.TrimSpace(req.TargetOrigin)
	if name == "" {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Field 'name' is required", nil)
		return
	}
	if targetOrigin == "" {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Field 'target_origin' is required", nil)
		return
	}

	sessionID := uuid.New().String()
	sess := &model.Session{
		ID:           sessionID,
		Name:         name,
		TargetOrigin: targetOrigin,
		Description:  strings.TrimSpace(req.Description),
	}

	if err := h.db.CreateSession(r.Context(), sess); err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create session", err.Error())
		return
	}

	scheme, host := getRequestHost(r)
	enrichSessionMetadata(sess, scheme, host)

	w.Header().Set("Location", fmt.Sprintf("/api/sessions/%s", sessionID))
	WriteJSON(w, http.StatusCreated, sess)
}

// GetSession handles GET /api/sessions/{id}.
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Session ID is required", nil)
		return
	}

	sess, err := h.db.GetSession(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get session", err.Error())
		return
	}
	if sess == nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Session with id '%s' not found", id), nil)
		return
	}

	scheme, host := getRequestHost(r)
	enrichSessionMetadata(sess, scheme, host)

	policySetting, _ := h.db.GetSessionPolicySetting(r.Context(), id)

	resp := SessionDetailResponse{
		Session:       *sess,
		PolicySetting: policySetting,
	}

	WriteJSON(w, http.StatusOK, resp)
}

// DeleteSession handles DELETE /api/sessions/{id}.
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Session ID is required", nil)
		return
	}

	deleted, err := h.db.DeleteSession(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete session", err.Error())
		return
	}
	if !deleted {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Session with id '%s' not found", id), nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getRequestHost(r *http.Request) (string, string) {
	scheme := "http"
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if r.TLS != nil {
		scheme = "https"
	}

	host := r.Host
	if fwdHost := r.Header.Get("X-Forwarded-Host"); fwdHost != "" {
		host = fwdHost
	}

	return scheme, host
}

func enrichSessionMetadata(s *model.Session, scheme, host string) {
	if scheme == "" {
		scheme = "http"
	}
	if host != "" {
		s.ReportURI = fmt.Sprintf("%s://%s/api/reports/%s", scheme, host, s.ID)
	} else {
		s.ReportURI = fmt.Sprintf("/api/reports/%s", s.ID)
	}
	s.HeaderSnippet = fmt.Sprintf("Content-Security-Policy-Report-Only: default-src 'none'; form-action 'none'; frame-ancestors 'none'; report-uri %s;", s.ReportURI)
}
