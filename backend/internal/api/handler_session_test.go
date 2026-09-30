package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"csp_report_backend/internal/model"
)

func TestSessionCRUD(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	router := SetupRouter(database)

	var createdID string

	t.Run("Create session successfully", func(t *testing.T) {
		body := `{
			"name": "Production App",
			"target_origin": "https://app.example.com",
			"description": "Main application production analysis"
		}`

		req, _ := http.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
		}

		location := rec.Header().Get("Location")
		if location == "" {
			t.Errorf("expected Location header in 201 response")
		}

		var sess model.Session
		if err := json.Unmarshal(rec.Body.Bytes(), &sess); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if sess.ID == "" {
			t.Errorf("expected non-empty session ID")
		}
		if sess.Name != "Production App" {
			t.Errorf("expected name 'Production App', got %s", sess.Name)
		}
		if sess.TargetOrigin != "https://app.example.com" {
			t.Errorf("expected target_origin 'https://app.example.com', got %s", sess.TargetOrigin)
		}
		if sess.ReportURI == "" {
			t.Errorf("expected generated ReportURI")
		}
		if sess.HeaderSnippet == "" {
			t.Errorf("expected generated HeaderSnippet")
		}

		createdID = sess.ID
	})

	t.Run("Create session validation error", func(t *testing.T) {
		body := `{"name": ""}`
		req, _ := http.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", rec.Code)
		}
	})

	t.Run("List sessions", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/sessions", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var sessions []model.Session
		if err := json.Unmarshal(rec.Body.Bytes(), &sessions); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(sessions) < 1 {
			t.Errorf("expected at least 1 session, got %d", len(sessions))
		}
	})

	t.Run("Get session details with policy settings", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/sessions/"+createdID, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp SessionDetailResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.ID != createdID {
			t.Errorf("expected session ID %s, got %s", createdID, resp.ID)
		}
		if resp.PolicySetting == nil {
			t.Errorf("expected default policy setting in response")
		} else if resp.PolicySetting.DefaultSrc != "'none'" {
			t.Errorf("expected default-src 'none', got %s", resp.PolicySetting.DefaultSrc)
		}
	})

	t.Run("Get non-existent session returns 404", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/sessions/non-existent-uuid", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found, got %d", rec.Code)
		}
	})

	t.Run("Delete session returns 204 and cascades", func(t *testing.T) {
		// Ingest a report first to verify cascading deletion
		report := &model.NormalizedViolation{
			DocumentURI:        "https://app.example.com",
			EffectiveDirective: "script-src",
			BlockedURI:         "https://cdn.example.com/bundle.js",
			OriginHost:         "https://cdn.example.com",
		}
		if err := database.IngestReport(context.Background(), report, createdID); err != nil {
			t.Fatalf("failed to ingest report: %v", err)
		}

		req, _ := http.NewRequest(http.MethodDelete, "/api/sessions/"+createdID, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204 No Content, got %d", rec.Code)
		}

		// Subsequent GET returns 404
		getReq, _ := http.NewRequest(http.MethodGet, "/api/sessions/"+createdID, nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)
		if getRec.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found after deletion, got %d", getRec.Code)
		}

		// Violations table should also be empty for this session
		groups, err := database.ListViolationGroups(context.Background(), createdID, "", "", "", "")
		if err != nil || len(groups) != 0 {
			t.Errorf("expected 0 violation groups after cascade deletion, got %d", len(groups))
		}
	})
}
