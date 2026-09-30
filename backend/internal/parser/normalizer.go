package parser

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"

	"csp_report_backend/internal/model"
)

// ParseReportPayload inspects incoming raw JSON bytes (CSPv2 object or Reporting API v1 array)
// and returns normalized violations.
func ParseReportPayload(payload []byte, targetOrigin string) ([]model.NormalizedViolation, error) {
	trimmed := strings.TrimSpace(string(payload))
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty payload")
	}

	var results []model.NormalizedViolation

	if strings.HasPrefix(trimmed, "[") {
		// Reporting API v1 format
		var items []model.ReportingAPIV1Item
		if err := json.Unmarshal(payload, &items); err != nil {
			return nil, fmt.Errorf("invalid Reporting API array JSON: %w", err)
		}
		for _, item := range items {
			nv := Normalize(
				item.Body.DocumentURL,
				item.Body.Referrer,
				item.Body.EffectiveDirective,
				item.Body.EffectiveDirective,
				item.Body.OriginalPolicy,
				item.Body.Disposition,
				item.Body.BlockedURL,
				item.Body.LineNumber,
				item.Body.ColumnNumber,
				item.Body.SourceFile,
				item.Body.StatusCode,
				item.Body.Sample,
				targetOrigin,
			)
			results = append(results, nv)
		}
		return results, nil
	}

	// Standard CSP report object: {"csp-report": { ... }}
	var std model.StandardCSPReportPayload
	if err := json.Unmarshal(payload, &std); err == nil && std.CSPReport.BlockedURI != "" {
		nv := Normalize(
			std.CSPReport.DocumentURI,
			std.CSPReport.Referrer,
			std.CSPReport.ViolatedDirective,
			std.CSPReport.EffectiveDirective,
			std.CSPReport.OriginalPolicy,
			std.CSPReport.Disposition,
			std.CSPReport.BlockedURI,
			std.CSPReport.LineNumber,
			std.CSPReport.ColumnNumber,
			std.CSPReport.SourceFile,
			std.CSPReport.StatusCode,
			std.CSPReport.ScriptSample,
			targetOrigin,
		)
		results = append(results, nv)
		return results, nil
	}

	// Try flat CSP report body without top-level "csp-report" key
	var flat model.StandardCSPReportBody
	if err := json.Unmarshal(payload, &flat); err == nil && flat.BlockedURI != "" {
		nv := Normalize(
			flat.DocumentURI,
			flat.Referrer,
			flat.ViolatedDirective,
			flat.EffectiveDirective,
			flat.OriginalPolicy,
			flat.Disposition,
			flat.BlockedURI,
			flat.LineNumber,
			flat.ColumnNumber,
			flat.SourceFile,
			flat.StatusCode,
			flat.ScriptSample,
			targetOrigin,
		)
		results = append(results, nv)
		return results, nil
	}

	return nil, fmt.Errorf("unrecognized CSP report JSON structure")
}

// Normalize processes violation attributes into standard origins, wildcards, and directive names.
func Normalize(
	documentURI, referrer, violatedDirective, effectiveDirective, originalPolicy, disposition,
	blockedURI string, lineNumber, columnNumber int, sourceFile string, statusCode int, scriptSample,
	targetOrigin string,
) model.NormalizedViolation {
	cleanDirective := cleanDirectiveName(effectiveDirective, violatedDirective)
	originHost, wildcard, isSelf := NormalizeBlockedURI(blockedURI, targetOrigin)

	return model.NormalizedViolation{
		DocumentURI:        documentURI,
		Referrer:           referrer,
		ViolatedDirective:  violatedDirective,
		EffectiveDirective: cleanDirective,
		OriginalPolicy:     originalPolicy,
		Disposition:        disposition,
		BlockedURI:         blockedURI,
		LineNumber:         lineNumber,
		ColumnNumber:       columnNumber,
		SourceFile:         sourceFile,
		StatusCode:         statusCode,
		ScriptSample:       scriptSample,
		OriginHost:         originHost,
		SuggestedWildcard:  wildcard,
		IsSelf:             isSelf,
	}
}

// NormalizeBlockedURI extracts origin, suggested wildcard, and matches against target origin.
func NormalizeBlockedURI(blockedURI, targetOrigin string) (originHost string, suggestedWildcard string, isSelf bool) {
	b := strings.TrimSpace(blockedURI)

	// Check special scheme keywords
	lower := strings.ToLower(b)
	switch {
	case strings.HasPrefix(lower, "data:"):
		return "data:", "", false
	case strings.HasPrefix(lower, "blob:"):
		return "blob:", "", false
	case strings.HasPrefix(lower, "filesystem:"):
		return "filesystem:", "", false
	case lower == "inline" || lower == "'inline'" || lower == "'unsafe-inline'":
		return "'unsafe-inline'", "", false
	case lower == "eval" || lower == "'eval'" || lower == "'unsafe-eval'":
		return "'unsafe-eval'", "", false
	case lower == "wasm-eval" || lower == "'wasm-eval'" || lower == "'wasm-unsafe-eval'":
		return "'wasm-unsafe-eval'", "", false
	case lower == "self" || lower == "'self'":
		return "'self'", "", true
	}

	// Handle relative paths or empty strings (always part of target origin 'self')
	if strings.HasPrefix(b, "/") || strings.HasPrefix(b, "./") || b == "" {
		if targetOrigin != "" {
			if targetURL, err := url.Parse(strings.TrimSpace(targetOrigin)); err == nil && targetURL.Host != "" {
				scheme := strings.ToLower(targetURL.Scheme)
				if scheme == "" {
					scheme = "https"
				}
				hostname := strings.ToLower(targetURL.Hostname())
				port := targetURL.Port()
				var oHost string
				if port != "" && port != "80" && port != "443" {
					oHost = fmt.Sprintf("%s://%s:%s", scheme, hostname, port)
				} else {
					oHost = fmt.Sprintf("%s://%s", scheme, hostname)
				}
				wildcard := deriveWildcard(scheme, hostname)
				return oHost, wildcard, true
			}
		}
		return "'self'", "", true
	}

	u, err := url.Parse(b)
	if err != nil || u.Host == "" {
		return b, "", false
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme == "" {
		scheme = "https"
	}

	hostname := strings.ToLower(u.Hostname())
	port := u.Port()

	if port != "" && port != "80" && port != "443" {
		originHost = fmt.Sprintf("%s://%s:%s", scheme, hostname, port)
	} else {
		originHost = fmt.Sprintf("%s://%s", scheme, hostname)
	}

	// Compare with session target origin
	if targetOrigin != "" {
		if targetURL, err := url.Parse(strings.TrimSpace(targetOrigin)); err == nil && targetURL.Host != "" {
			targetHost := strings.ToLower(targetURL.Hostname())
			targetScheme := strings.ToLower(targetURL.Scheme)
			if targetScheme == "" {
				targetScheme = "https"
			}
			if hostname == targetHost && (scheme == targetScheme || targetScheme == "") {
				if targetURL.Port() == port {
					isSelf = true
				}
			}
		}
	}

	// Compute wildcard suggestion
	suggestedWildcard = deriveWildcard(scheme, hostname)

	return originHost, suggestedWildcard, isSelf
}

func deriveWildcard(scheme, hostname string) string {
	// IP address or localhost should not have wildcards
	if net.ParseIP(hostname) != nil || hostname == "localhost" || strings.HasSuffix(hostname, ".localhost") {
		return ""
	}

	parts := strings.Split(hostname, ".")
	if len(parts) <= 2 {
		// e.g. "domain.com" or "example.org"
		return fmt.Sprintf("%s://*.%s", scheme, hostname)
	}

	// Handle multi-part public suffixes (e.g., .co.uk, .com.es, .org.uk, .gov.uk)
	var baseDomain string
	twoPartTLDs := []string{".co.uk", ".org.uk", ".com.es", ".nom.es", ".org.es", ".com.br", ".com.mx", ".co.jp"}
	isTwoPart := false
	for _, tld := range twoPartTLDs {
		if strings.HasSuffix(hostname, tld) {
			isTwoPart = true
			break
		}
	}

	if isTwoPart && len(parts) >= 3 {
		baseDomain = strings.Join(parts[len(parts)-3:], ".")
	} else {
		baseDomain = strings.Join(parts[len(parts)-2:], ".")
	}

	return fmt.Sprintf("%s://*.%s", scheme, baseDomain)
}

func cleanDirectiveName(effectiveDirective, violatedDirective string) string {
	d := effectiveDirective
	if d == "" {
		d = violatedDirective
	}
	d = strings.TrimSpace(d)
	// Some browsers report "script-src-elem" or "script-src-attr", or include directives like "style-src 'none'"
	if idx := strings.Index(d, " "); idx != -1 {
		d = d[:idx]
	}
	return strings.ToLower(d)
}
