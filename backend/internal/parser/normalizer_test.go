package parser

import (
	"testing"
)

func TestNormalizeBlockedURI(t *testing.T) {
	targetOrigin := "https://app.example.com"

	tests := []struct {
		name              string
		blockedURI        string
		wantOrigin        string
		wantWildcard      string
		wantIsSelf        bool
	}{
		{
			name:         "API endpoint",
			blockedURI:   "https://api.service.example.com/v1/resources",
			wantOrigin:   "https://api.service.example.com",
			wantWildcard: "https://*.example.com",
			wantIsSelf:   false,
		},
		{
			name:         "Same host as target origin (is_self)",
			blockedURI:   "https://app.example.com/polyfills-B6TNHZQ6.js",
			wantOrigin:   "https://app.example.com",
			wantWildcard: "https://*.example.com",
			wantIsSelf:   true,
		},
		{
			name:         "Custom port URL",
			blockedURI:   "http://localhost:8080/api/reports/123",
			wantOrigin:   "http://localhost:8080",
			wantWildcard: "",
			wantIsSelf:   false,
		},
		{
			name:         "Data URI",
			blockedURI:   "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
			wantOrigin:   "data:",
			wantWildcard: "",
			wantIsSelf:   false,
		},
		{
			name:         "Blob URI",
			blockedURI:   "blob:https://app.example.com/uuid",
			wantOrigin:   "blob:",
			wantWildcard: "",
			wantIsSelf:   false,
		},
		{
			name:         "Inline script violation",
			blockedURI:   "inline",
			wantOrigin:   "'unsafe-inline'",
			wantWildcard: "",
			wantIsSelf:   false,
		},
		{
			name:         "Eval violation",
			blockedURI:   "eval",
			wantOrigin:   "'unsafe-eval'",
			wantWildcard: "",
			wantIsSelf:   false,
		},
		{
			name:         "Multi-part TLD (.co.uk)",
			blockedURI:   "https://api.service.company.co.uk/v1/users",
			wantOrigin:   "https://api.service.company.co.uk",
			wantWildcard: "https://*.company.co.uk",
			wantIsSelf:   false,
		},
		{
			name:         "Relative root path (is_self)",
			blockedURI:   "/assets/main.js",
			wantOrigin:   "https://app.example.com",
			wantWildcard: "https://*.example.com",
			wantIsSelf:   true,
		},
		{
			name:         "Relative dot path (is_self)",
			blockedURI:   "./style.css",
			wantOrigin:   "https://app.example.com",
			wantWildcard: "https://*.example.com",
			wantIsSelf:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOrigin, gotWildcard, gotIsSelf := NormalizeBlockedURI(tt.blockedURI, targetOrigin)
			if gotOrigin != tt.wantOrigin {
				t.Errorf("NormalizeBlockedURI() gotOrigin = %v, want %v", gotOrigin, tt.wantOrigin)
			}
			if gotWildcard != tt.wantWildcard {
				t.Errorf("NormalizeBlockedURI() gotWildcard = %v, want %v", gotWildcard, tt.wantWildcard)
			}
			if gotIsSelf != tt.wantIsSelf {
				t.Errorf("NormalizeBlockedURI() gotIsSelf = %v, want %v", gotIsSelf, tt.wantIsSelf)
			}
		})
	}
}

func TestParseReportPayload_Standard(t *testing.T) {
	rawPayload := []byte(`{
	  "csp-report": {
	    "document-uri": "https://app.example.com/dashboard",
	    "referrer": "https://app.example.com/",
	    "violated-directive": "connect-src",
	    "effective-directive": "connect-src",
	    "original-policy": "default-src 'none'; form-action 'none'; frame-ancestors 'none'; report-uri http://127.0.0.1:8080/91998b61-f38e-4b38-9f42-ea68780e8a77",
	    "disposition": "report",
	    "blocked-uri": "https://api.service.example.com/v1/resources",
	    "line-number": 1,
	    "column-number": 25897,
	    "source-file": "https://app.example.com/bundle.js",
	    "status-code": 200,
	    "script-sample": ""
	  }
	}`)

	violations, err := ParseReportPayload(rawPayload, "https://app.example.com")
	if err != nil {
		t.Fatalf("ParseReportPayload failed: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.EffectiveDirective != "connect-src" {
		t.Errorf("expected connect-src, got %s", v.EffectiveDirective)
	}
	if v.OriginHost != "https://api.service.example.com" {
		t.Errorf("expected https://api.service.example.com, got %s", v.OriginHost)
	}
	if v.SuggestedWildcard != "https://*.example.com" {
		t.Errorf("expected https://*.example.com, got %s", v.SuggestedWildcard)
	}
	if v.IsSelf {
		t.Errorf("expected is_self false, got true")
	}
	if v.SourceFile != "https://app.example.com/bundle.js" {
		t.Errorf("expected bundle source file, got %s", v.SourceFile)
	}
}

func TestParseReportPayload_ReportingAPI(t *testing.T) {
	rawPayload := []byte(`[
	  {
	    "type": "csp-violation",
	    "age": 10,
	    "url": "https://app.example.com/dashboard",
	    "user_agent": "Mozilla/5.0",
	    "body": {
	      "documentURL": "https://app.example.com/dashboard",
	      "referrer": "https://app.example.com/",
	      "blockedURL": "https://cdn.jsdelivr.net/npm/chart.js",
	      "effectiveDirective": "script-src-elem",
	      "originalPolicy": "default-src 'none'",
	      "sourceFile": "https://app.example.com/app.js",
	      "sample": "",
	      "disposition": "report",
	      "statusCode": 200,
	      "lineNumber": 42,
	      "columnNumber": 12
	    }
	  }
	]`)

	violations, err := ParseReportPayload(rawPayload, "https://app.example.com")
	if err != nil {
		t.Fatalf("ParseReportPayload failed: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.OriginHost != "https://cdn.jsdelivr.net" {
		t.Errorf("expected https://cdn.jsdelivr.net, got %s", v.OriginHost)
	}
	if v.SuggestedWildcard != "https://*.jsdelivr.net" {
		t.Errorf("expected https://*.jsdelivr.net, got %s", v.SuggestedWildcard)
	}
}
