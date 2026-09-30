package model

import "time"

// Session represents an analysis project targeting a specific web origin.
type Session struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	TargetOrigin      string    `json:"target_origin"`
	Description       string    `json:"description,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	TotalReports      int       `json:"total_reports,omitempty"`
	PendingViolations int       `json:"pending_violations,omitempty"`
	ApprovedRules     int       `json:"approved_rules,omitempty"`
	SelfViolations    int       `json:"self_violations,omitempty"`
	ReportURI         string    `json:"report_uri,omitempty"`
	HeaderSnippet     string    `json:"header_snippet,omitempty"`
}

// RawReport represents an unedited violation report received from a browser.
type RawReport struct {
	ID                 int64     `json:"id"`
	SessionID          string    `json:"session_id"`
	DocumentURI        string    `json:"document_uri"`
	Referrer           string    `json:"referrer"`
	ViolatedDirective  string    `json:"violated_directive"`
	EffectiveDirective string    `json:"effective_directive"`
	OriginalPolicy     string    `json:"original_policy"`
	Disposition        string    `json:"disposition"`
	BlockedURI         string    `json:"blocked_uri"`
	LineNumber         int       `json:"line_number"`
	ColumnNumber       int       `json:"column_number"`
	SourceFile         string    `json:"source_file"`
	StatusCode         int       `json:"status_code"`
	ScriptSample       string    `json:"script_sample"`
	CreatedAt          time.Time `json:"created_at"`
}

// ViolationGroup aggregates violations by directive and normalized origin host.
type ViolationGroup struct {
	ID                int64     `json:"id"`
	SessionID         string    `json:"session_id"`
	Directive         string    `json:"directive"`
	OriginHost        string    `json:"origin_host"`
	SuggestedWildcard string    `json:"suggested_wildcard,omitempty"`
	IsSelf            bool      `json:"is_self"`
	Count             int       `json:"count"`
	LastSeenAt        time.Time `json:"last_seen_at"`
	Status            string    `json:"status"` // pending | approved_origin | approved_wildcard | approved_self | rejected | ignored
}

// SessionPolicySetting stores configuration toggles for compiling the final CSP.
type SessionPolicySetting struct {
	SessionID               string `json:"session_id"`
	DefaultSrc              string `json:"default_src"`
	FormAction              string `json:"form_action"`
	FrameAncestors          string `json:"frame_ancestors"`
	UpgradeInsecureRequests bool   `json:"upgrade_insecure_requests"`
	BlockAllMixedContent    bool   `json:"block_all_mixed_content"`
	ReportOnly              bool   `json:"report_only"`
	CustomDirectives        string `json:"custom_directives,omitempty"`
}

// StandardCSPReportBody matches the standard legacy W3C CSP Report format.
type StandardCSPReportBody struct {
	DocumentURI        string `json:"document-uri"`
	Referrer           string `json:"referrer"`
	ViolatedDirective  string `json:"violated-directive"`
	EffectiveDirective string `json:"effective-directive"`
	OriginalPolicy     string `json:"original-policy"`
	Disposition        string `json:"disposition"`
	BlockedURI         string `json:"blocked-uri"`
	LineNumber         int    `json:"line-number"`
	ColumnNumber       int    `json:"column-number"`
	SourceFile         string `json:"source-file"`
	StatusCode         int    `json:"status-code"`
	ScriptSample       string `json:"script-sample"`
}

// StandardCSPReportPayload wraps the top-level "csp-report" key.
type StandardCSPReportPayload struct {
	CSPReport StandardCSPReportBody `json:"csp-report"`
}

// ReportingAPIV1Item matches W3C Reporting API v1 item for csp-violation.
type ReportingAPIV1Item struct {
	Type string `json:"type"`
	Age  int    `json:"age"`
	URL  string `json:"url"`
	UserAgent string `json:"user_agent"`
	Body struct {
		DocumentURL        string `json:"documentURL"`
		Referrer           string `json:"referrer"`
		BlockedURL         string `json:"blockedURL"`
		EffectiveDirective string `json:"effectiveDirective"`
		OriginalPolicy     string `json:"originalPolicy"`
		SourceFile         string `json:"sourceFile"`
		Sample             string `json:"sample"`
		Disposition        string `json:"disposition"`
		StatusCode         int    `json:"statusCode"`
		LineNumber         int    `json:"lineNumber"`
		ColumnNumber       int    `json:"columnNumber"`
	} `json:"body"`
}

// NormalizedViolation is intermediate data extracted from any CSP report variant.
type NormalizedViolation struct {
	DocumentURI        string
	Referrer           string
	ViolatedDirective  string
	EffectiveDirective string
	OriginalPolicy     string
	Disposition        string
	BlockedURI         string
	LineNumber         int
	ColumnNumber       int
	SourceFile         string
	StatusCode         int
	ScriptSample       string
	OriginHost         string
	SuggestedWildcard  string
	IsSelf             bool
}

// ErrorDetail provides structured error details.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ErrorResponse represents the standard REST error envelope.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
