package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"csp_report_backend/internal/generator"
	"csp_report_backend/internal/model"
)

func TestViolationsAndPolicyWorkflow(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	router := SetupRouter(database)

	sess := &model.Session{
		ID:           "workflow-session-1",
		Name:         "Portal App",
		TargetOrigin: "https://app.example.com",
	}
	if err := database.CreateSession(context.Background(), sess); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// 1. Ingest report
	reportPayload := `{
		"csp-report": {
			"document-uri": "https://app.example.com/dashboard",
			"blocked-uri": "https://api.example.com/v1/resources",
			"effective-directive": "connect-src",
			"line-number": 42,
			"source-file": "https://app.example.com/app.js"
		}
	}`
	req1, _ := http.NewRequest(http.MethodPost, "/api/reports/workflow-session-1", bytes.NewBufferString(reportPayload))
	req1.Header.Set("Content-Type", "application/csp-report")
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content on ingest, got %d", rec1.Code)
	}

	// Ingest a second identical report to verify count increment
	req2, _ := http.NewRequest(http.MethodPost, "/api/reports/workflow-session-1", bytes.NewBufferString(reportPayload))
	req2.Header.Set("Content-Type", "application/csp-report")
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content on 2nd ingest, got %d", rec2.Code)
	}

	var violationID int64

	t.Run("List violations and verify count", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/sessions/workflow-session-1/violations?directive=connect-src", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var violations []model.ViolationGroup
		if err := json.Unmarshal(rec.Body.Bytes(), &violations); err != nil {
			t.Fatalf("failed to decode violations: %v", err)
		}

		if len(violations) != 1 {
			t.Fatalf("expected 1 violation, got %d", len(violations))
		}
		if violations[0].Count != 2 {
			t.Errorf("expected count 2, got %d", violations[0].Count)
		}
		if violations[0].SuggestedWildcard != "https://*.example.com" {
			t.Errorf("expected wildcard https://*.example.com, got %s", violations[0].SuggestedWildcard)
		}

		violationID = violations[0].ID
	})

	t.Run("Get violation samples", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/sessions/workflow-session-1/violations/%d/samples", violationID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var samples []model.RawReport
		if err := json.Unmarshal(rec.Body.Bytes(), &samples); err != nil {
			t.Fatalf("failed to decode samples: %v", err)
		}

		if len(samples) != 2 {
			t.Fatalf("expected 2 raw report samples, got %d", len(samples))
		}
		if samples[0].LineNumber != 42 {
			t.Errorf("expected line number 42, got %d", samples[0].LineNumber)
		}
	})

	t.Run("Patch violation status to approved_wildcard", func(t *testing.T) {
		body := `{"status": "approved_wildcard"}`
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/sessions/workflow-session-1/violations/%d", violationID), bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		var updated model.ViolationGroup
		if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
			t.Fatalf("failed to decode updated violation: %v", err)
		}

		if updated.Status != "approved_wildcard" {
			t.Errorf("expected status 'approved_wildcard', got %s", updated.Status)
		}
	})

	t.Run("Bulk update violations", func(t *testing.T) {
		body := fmt.Sprintf(`{"violation_ids": [%d], "status": "approved_origin"}`, violationID)
		req, _ := http.NewRequest(http.MethodPost, "/api/sessions/workflow-session-1/violations/bulk", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var resp map[string]int
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode bulk resp: %v", err)
		}
		if resp["updated"] != 1 {
			t.Errorf("expected 1 updated, got %d", resp["updated"])
		}
	})

	t.Run("Update policy settings", func(t *testing.T) {
		body := `{
			"default_src": "'self'",
			"form_action": "'self'",
			"frame_ancestors": "'none'",
			"upgrade_insecure_requests": true,
			"report_only": false
		}`
		req, _ := http.NewRequest(http.MethodPut, "/api/sessions/workflow-session-1/settings", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var setting model.SessionPolicySetting
		if err := json.Unmarshal(rec.Body.Bytes(), &setting); err != nil {
			t.Fatalf("failed to decode settings: %v", err)
		}
		if setting.DefaultSrc != "'self'" {
			t.Errorf("expected default_src 'self', got %s", setting.DefaultSrc)
		}
		if !setting.UpgradeInsecureRequests {
			t.Errorf("expected upgrade_insecure_requests true")
		}
	})

	t.Run("Export compiled policy in JSON and raw formats", func(t *testing.T) {
		// All formats JSON
		req, _ := http.NewRequest(http.MethodGet, "/api/sessions/workflow-session-1/policy", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var out generator.PolicyOutput
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("failed to decode policy output: %v", err)
		}

		if !strings.Contains(out.Raw, "connect-src https://api.example.com") {
			t.Errorf("expected approved origin in policy raw: %s", out.Raw)
		}
		if !strings.Contains(out.Raw, "upgrade-insecure-requests") {
			t.Errorf("expected upgrade-insecure-requests in raw: %s", out.Raw)
		}

		// Raw format query param
		reqRaw, _ := http.NewRequest(http.MethodGet, "/api/sessions/workflow-session-1/export?format=raw", nil)
		recRaw := httptest.NewRecorder()
		router.ServeHTTP(recRaw, reqRaw)

		if recRaw.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", recRaw.Code)
		}
		if recRaw.Body.String() != out.Raw {
			t.Errorf("raw export mismatch. Expected %s, got %s", out.Raw, recRaw.Body.String())
		}

		// Nginx format query param
		reqNginx, _ := http.NewRequest(http.MethodGet, "/api/sessions/workflow-session-1/export?format=nginx", nil)
		recNginx := httptest.NewRecorder()
		router.ServeHTTP(recNginx, reqNginx)

		if recNginx.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", recNginx.Code)
		}
		if recNginx.Body.String() != out.Nginx {
			t.Errorf("nginx export mismatch. Expected %s, got %s", out.Nginx, recNginx.Body.String())
		}
	})

	t.Run("Self origin filtering and bulk approve self", func(t *testing.T) {
		selfPayload := `{
			"csp-report": {
				"document-uri": "https://app.example.com/dashboard",
				"blocked-uri": "/static/js/main.chunk.js",
				"effective-directive": "script-src",
				"line-number": 1
			}
		}`
		reqSelf, _ := http.NewRequest(http.MethodPost, "/api/reports/workflow-session-1", bytes.NewBufferString(selfPayload))
		reqSelf.Header.Set("Content-Type", "application/csp-report")
		recSelf := httptest.NewRecorder()
		router.ServeHTTP(recSelf, reqSelf)
		if recSelf.Code != http.StatusNoContent {
			t.Fatalf("expected 204 No Content on self ingest, got %d", recSelf.Code)
		}

		reqListSelf, _ := http.NewRequest(http.MethodGet, "/api/sessions/workflow-session-1/violations?source=self", nil)
		recListSelf := httptest.NewRecorder()
		router.ServeHTTP(recListSelf, reqListSelf)
		if recListSelf.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", recListSelf.Code)
		}
		var selfGroups []model.ViolationGroup
		if err := json.Unmarshal(recListSelf.Body.Bytes(), &selfGroups); err != nil {
			t.Fatalf("failed to decode self groups: %v", err)
		}
		if len(selfGroups) != 1 || !selfGroups[0].IsSelf {
			t.Fatalf("expected 1 self violation group, got %d", len(selfGroups))
		}

		reqListTP, _ := http.NewRequest(http.MethodGet, "/api/sessions/workflow-session-1/violations?source=third_party", nil)
		recListTP := httptest.NewRecorder()
		router.ServeHTTP(recListTP, reqListTP)
		if recListTP.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", recListTP.Code)
		}
		var tpGroups []model.ViolationGroup
		if err := json.Unmarshal(recListTP.Body.Bytes(), &tpGroups); err != nil {
			t.Fatalf("failed to decode tp groups: %v", err)
		}
		for _, g := range tpGroups {
			if g.IsSelf {
				t.Fatalf("unexpected self violation in third_party filter: %+v", g)
			}
		}

		reqApproveSelf, _ := http.NewRequest(http.MethodPost, "/api/sessions/workflow-session-1/violations/approve-self", nil)
		recApproveSelf := httptest.NewRecorder()
		router.ServeHTTP(recApproveSelf, reqApproveSelf)
		if recApproveSelf.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recApproveSelf.Code, recApproveSelf.Body.String())
		}
		var approveResp map[string]int
		if err := json.Unmarshal(recApproveSelf.Body.Bytes(), &approveResp); err != nil {
			t.Fatalf("failed to decode approve-self resp: %v", err)
		}
		if approveResp["updated"] < 1 {
			t.Errorf("expected at least 1 updated for approve-self, got %d", approveResp["updated"])
		}
	})
}
