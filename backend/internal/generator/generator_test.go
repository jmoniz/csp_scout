package generator

import (
	"strings"
	"testing"

	"csp_report_backend/internal/model"
)

func TestGeneratePolicy(t *testing.T) {
	t.Run("Default settings with no violations", func(t *testing.T) {
		setting := &model.SessionPolicySetting{
			SessionID:      "test-1",
			DefaultSrc:     "'none'",
			FormAction:     "'none'",
			FrameAncestors: "'none'",
			ReportOnly:     false,
		}

		out := GeneratePolicy(setting, nil)

		expectedRaw := "default-src 'none'; form-action 'none'; frame-ancestors 'none';"
		if out.Raw != expectedRaw {
			t.Errorf("expected raw policy %q, got %q", expectedRaw, out.Raw)
		}
		if out.HeaderName != "Content-Security-Policy" {
			t.Errorf("expected header name Content-Security-Policy, got %s", out.HeaderName)
		}
		if !strings.Contains(out.Header, "Content-Security-Policy: "+expectedRaw) {
			t.Errorf("unexpected header output: %s", out.Header)
		}
		if !strings.Contains(out.Nginx, `add_header Content-Security-Policy "`+expectedRaw+`";`) {
			t.Errorf("unexpected nginx output: %s", out.Nginx)
		}
	})

	t.Run("ReportOnly flag generates Report-Only header", func(t *testing.T) {
		setting := &model.SessionPolicySetting{
			SessionID:  "test-2",
			DefaultSrc: "'self'",
			ReportOnly: true,
		}

		out := GeneratePolicy(setting, nil)
		if out.HeaderName != "Content-Security-Policy-Report-Only" {
			t.Errorf("expected Content-Security-Policy-Report-Only, got %s", out.HeaderName)
		}
		if !strings.HasPrefix(out.Header, "Content-Security-Policy-Report-Only:") {
			t.Errorf("header does not start with report only name: %s", out.Header)
		}
	})

	t.Run("Approved sources drop 'none' and quote keywords properly", func(t *testing.T) {
		setting := &model.SessionPolicySetting{
			SessionID:      "test-3",
			DefaultSrc:     "'none'",
			FormAction:     "'none'",
			FrameAncestors: "'none'",
		}

		violations := []model.ViolationGroup{
			{
				Directive:  "script-src",
				OriginHost: "'self'",
				Status:     "approved_self",
			},
			{
				Directive:  "script-src",
				OriginHost: "'unsafe-inline'",
				Status:     "approved_origin",
			},
			{
				Directive:  "connect-src",
				OriginHost: "https://api.example.com",
				Status:     "approved_origin",
			},
			{
				Directive:  "img-src",
				OriginHost: "data:",
				Status:     "approved_origin",
			},
		}

		out := GeneratePolicy(setting, violations)

		if !strings.Contains(out.Raw, "script-src 'self' 'unsafe-inline'") {
			t.Errorf("script-src not formatted properly: %s", out.Raw)
		}
		if !strings.Contains(out.Raw, "connect-src https://api.example.com") {
			t.Errorf("connect-src not formatted properly: %s", out.Raw)
		}
		if !strings.Contains(out.Raw, "img-src data:") {
			t.Errorf("img-src not formatted properly: %s", out.Raw)
		}
	})

	t.Run("Wildcard rule cleans up redundant exact origins", func(t *testing.T) {
		setting := &model.SessionPolicySetting{
			SessionID:  "test-4",
			DefaultSrc: "'self'",
		}

		violations := []model.ViolationGroup{
			{
				Directive:         "connect-src",
				OriginHost:        "https://api-dev.example.com",
				SuggestedWildcard: "https://*.example.com",
				Status:            "approved_wildcard",
			},
			{
				Directive:  "connect-src",
				OriginHost: "https://api-dev.example.com",
				Status:     "approved_origin",
			},
			{
				Directive:  "connect-src",
				OriginHost: "https://auth.example.com",
				Status:     "approved_origin",
			},
			{
				Directive:  "connect-src",
				OriginHost: "https://other-domain.com",
				Status:     "approved_origin",
			},
		}

		out := GeneratePolicy(setting, violations)

		// connect-src should contain https://*.example.com and https://other-domain.com
		// but should NOT contain https://api-dev.example.com or https://auth.example.com
		if !strings.Contains(out.Raw, "https://*.example.com") {
			t.Errorf("expected wildcard in connect-src: %s", out.Raw)
		}
		if !strings.Contains(out.Raw, "https://other-domain.com") {
			t.Errorf("expected https://other-domain.com in connect-src: %s", out.Raw)
		}
		if strings.Contains(out.Raw, "https://api-dev.example.com") {
			t.Errorf("expected redundant exact origin to be removed: %s", out.Raw)
		}
		if strings.Contains(out.Raw, "https://auth.example.com") {
			t.Errorf("expected redundant auth origin to be removed: %s", out.Raw)
		}
	})

	t.Run("Boolean directives and custom directives", func(t *testing.T) {
		setting := &model.SessionPolicySetting{
			SessionID:               "test-5",
			DefaultSrc:              "'self'",
			UpgradeInsecureRequests: true,
			BlockAllMixedContent:    true,
			CustomDirectives:        "sandbox allow-scripts; report-to endpoint",
		}

		out := GeneratePolicy(setting, nil)

		if !strings.Contains(out.Raw, "upgrade-insecure-requests") {
			t.Errorf("expected upgrade-insecure-requests in output: %s", out.Raw)
		}
		if !strings.Contains(out.Raw, "block-all-mixed-content") {
			t.Errorf("expected block-all-mixed-content in output: %s", out.Raw)
		}
		if !strings.Contains(out.Raw, "sandbox allow-scripts") {
			t.Errorf("expected custom sandbox directive: %s", out.Raw)
		}
		if !strings.Contains(out.Raw, "report-to endpoint") {
			t.Errorf("expected custom report-to directive: %s", out.Raw)
		}
	})
}
