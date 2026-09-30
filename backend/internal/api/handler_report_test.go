package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"csp_report_backend/internal/db"
	"csp_report_backend/internal/model"
)

func setupTestDB(t *testing.T) *db.DB {
	database, err := db.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open test in-memory db: %v", err)
	}
	return database
}

func TestHealthCheck(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	router := SetupRouter(database)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %s", resp["status"])
	}
}

func TestReportIngestion(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	// Seed session
	sess := &model.Session{
		ID:           "test-sess-123",
		Name:         "Portal Test",
		TargetOrigin: "https://app.example.com",
	}
	if err := database.CreateSession(context.Background(), sess); err != nil {
		t.Fatalf("failed to seed session: %v", err)
	}

	router := SetupRouter(database)

	t.Run("CORS preflight OPTIONS on /api/reports/{id}", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodOptions, "/api/reports/test-sess-123", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204 No Content, got %d", rec.Code)
		}
		if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("expected Access-Control-Allow-Origin *, got %s", origin)
		}
	})

	t.Run("CORS preflight OPTIONS on /{id} alias", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodOptions, "/test-sess-123", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204 No Content, got %d", rec.Code)
		}
	})

	t.Run("Standard CSP Report payload", func(t *testing.T) {
		payload := `{
			"csp-report": {
				"document-uri": "https://app.example.com/dashboard",
				"referrer": "",
				"violated-directive": "connect-src",
				"effective-directive": "connect-src",
				"original-policy": "default-src 'self'",
				"disposition": "report",
				"blocked-uri": "https://api.example.com/v1/resources",
				"status-code": 200
			}
		}`

		req, _ := http.NewRequest(http.MethodPost, "/api/reports/test-sess-123", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/csp-report")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204 No Content, got %d: %s", rec.Code, rec.Body.String())
		}

		groups, err := database.ListViolationGroups(context.Background(), "test-sess-123", "", "", "", "")
		if err != nil || len(groups) != 1 {
			t.Fatalf("expected 1 violation group, got %d: %v", len(groups), err)
		}
		if groups[0].OriginHost != "https://api.example.com" {
			t.Errorf("expected origin https://api.example.com, got %s", groups[0].OriginHost)
		}
		if groups[0].Count != 1 {
			t.Errorf("expected count 1, got %d", groups[0].Count)
		}
	})

	t.Run("Reporting API v1 payload array", func(t *testing.T) {
		payload := `[
			{
				"type": "csp-violation",
				"age": 10,
				"url": "https://app.example.com/login",
				"user_agent": "Mozilla/5.0",
				"body": {
					"documentURL": "https://app.example.com/login",
					"blockedURL": "https://fonts.googleapis.com/css2?family=Roboto",
					"effectiveDirective": "style-src",
					"originalPolicy": "default-src 'self'"
				}
			}
		]`

		req, _ := http.NewRequest(http.MethodPost, "/api/reports/test-sess-123", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/reports+json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204 No Content, got %d: %s", rec.Code, rec.Body.String())
		}

		groups, err := database.ListViolationGroups(context.Background(), "test-sess-123", "style-src", "", "", "")
		if err != nil || len(groups) != 1 {
			t.Fatalf("expected 1 style-src group, got %d", len(groups))
		}
		if groups[0].OriginHost != "https://fonts.googleapis.com" {
			t.Errorf("expected https://fonts.googleapis.com, got %s", groups[0].OriginHost)
		}
	})

	t.Run("Root alias /{sessionId}", func(t *testing.T) {
		payload := `{
			"csp-report": {
				"document-uri": "https://app.example.com/profile",
				"blocked-uri": "https://img.example.com/logo.png",
				"effective-directive": "img-src"
			}
		}`

		req, _ := http.NewRequest(http.MethodPost, "/test-sess-123", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204 No Content on root alias, got %d", rec.Code)
		}
	})

	t.Run("Report to non-existent session returns 404", func(t *testing.T) {
		payload := `{"csp-report": {"blocked-uri": "https://evil.com", "effective-directive": "script-src"}}`
		req, _ := http.NewRequest(http.MethodPost, "/api/reports/non-existent-session-id", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found, got %d", rec.Code)
		}

		var errResp model.ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp.Error.Code != "NOT_FOUND" {
			t.Errorf("expected error code NOT_FOUND, got %s", errResp.Error.Code)
		}
	})

	t.Run("Malformed JSON payload returns 400", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/api/reports/test-sess-123", bytes.NewBufferString("not-valid-json{"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", rec.Code)
		}
	})
}
