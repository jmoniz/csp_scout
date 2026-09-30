package generator

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"csp_report_backend/internal/model"
)

// PolicyOutput contains the generated CSP in various standard formats.
type PolicyOutput struct {
	Raw        string `json:"raw"`
	HeaderName string `json:"header_name"`
	Header     string `json:"header"`
	Meta       string `json:"meta"`
	Nginx      string `json:"nginx"`
	Apache     string `json:"apache"`
}

// Ordered directive list for consistent and clean policy generation.
var directiveOrder = []string{
	"default-src",
	"script-src",
	"script-src-elem",
	"script-src-attr",
	"style-src",
	"style-src-elem",
	"style-src-attr",
	"img-src",
	"font-src",
	"connect-src",
	"media-src",
	"object-src",
	"child-src",
	"frame-src",
	"worker-src",
	"manifest-src",
	"form-action",
	"frame-ancestors",
	"base-uri",
}

var standardKeywords = map[string]string{
	"none":                 "'none'",
	"'none'":               "'none'",
	"self":                 "'self'",
	"'self'":               "'self'",
	"unsafe-inline":        "'unsafe-inline'",
	"'unsafe-inline'":      "'unsafe-inline'",
	"unsafe-eval":          "'unsafe-eval'",
	"'unsafe-eval'":        "'unsafe-eval'",
	"wasm-unsafe-eval":     "'wasm-unsafe-eval'",
	"'wasm-unsafe-eval'":   "'wasm-unsafe-eval'",
	"strict-dynamic":       "'strict-dynamic'",
	"'strict-dynamic'":     "'strict-dynamic'",
	"report-sample":        "'report-sample'",
	"'report-sample'":      "'report-sample'",
	"unsafe-hashes":        "'unsafe-hashes'",
	"'unsafe-hashes'":      "'unsafe-hashes'",
}

// GeneratePolicy compiles approved violations and settings into CSP formats.
func GeneratePolicy(setting *model.SessionPolicySetting, approvedViolations []model.ViolationGroup) PolicyOutput {
	if setting == nil {
		setting = &model.SessionPolicySetting{
			DefaultSrc:     "'none'",
			FormAction:     "'none'",
			FrameAncestors: "'none'",
		}
	}

	directiveMap := make(map[string][]string)

	// Apply base setting defaults if specified
	if setting.DefaultSrc != "" {
		directiveMap["default-src"] = parseSourceList(setting.DefaultSrc)
	}
	if setting.FormAction != "" {
		directiveMap["form-action"] = parseSourceList(setting.FormAction)
	}
	if setting.FrameAncestors != "" {
		directiveMap["frame-ancestors"] = parseSourceList(setting.FrameAncestors)
	}

	// Add approved violations
	for _, v := range approvedViolations {
		dir := strings.TrimSpace(strings.ToLower(v.Directive))
		if dir == "" {
			continue
		}

		var src string
		switch v.Status {
		case "approved_origin":
			src = v.OriginHost
		case "approved_wildcard":
			if v.SuggestedWildcard != "" {
				src = v.SuggestedWildcard
			} else {
				src = v.OriginHost
			}
		case "approved_self":
			src = "'self'"
		default:
			continue
		}

		cleanSrc := normalizeSourceToken(src)
		if cleanSrc != "" {
			directiveMap[dir] = append(directiveMap[dir], cleanSrc)
		}
	}

	// Build directive entries in standard order
	var compiledDirectives []string

	// Directives from our standard list
	for _, dir := range directiveOrder {
		sources, exists := directiveMap[dir]
		if !exists {
			continue
		}
		optimized := optimizeSources(sources)
		if len(optimized) > 0 {
			compiledDirectives = append(compiledDirectives, fmt.Sprintf("%s %s", dir, strings.Join(optimized, " ")))
		}
	}

	// Directives not in standard order
	var extraDirectives []string
	for dir := range directiveMap {
		if !isInDirectiveOrder(dir) {
			extraDirectives = append(extraDirectives, dir)
		}
	}
	sort.Strings(extraDirectives)
	for _, dir := range extraDirectives {
		sources := directiveMap[dir]
		optimized := optimizeSources(sources)
		if len(optimized) > 0 {
			compiledDirectives = append(compiledDirectives, fmt.Sprintf("%s %s", dir, strings.Join(optimized, " ")))
		}
	}

	// Boolean directives
	if setting.UpgradeInsecureRequests {
		compiledDirectives = append(compiledDirectives, "upgrade-insecure-requests")
	}
	if setting.BlockAllMixedContent {
		compiledDirectives = append(compiledDirectives, "block-all-mixed-content")
	}

	// Custom directives append
	if strings.TrimSpace(setting.CustomDirectives) != "" {
		customParts := strings.Split(setting.CustomDirectives, ";")
		for _, part := range customParts {
			p := strings.TrimSpace(part)
			if p != "" {
				compiledDirectives = append(compiledDirectives, p)
			}
		}
	}

	raw := strings.Join(compiledDirectives, "; ")
	if raw != "" {
		raw += ";"
	}

	headerName := "Content-Security-Policy"
	if setting.ReportOnly {
		headerName = "Content-Security-Policy-Report-Only"
	}

	return PolicyOutput{
		Raw:        raw,
		HeaderName: headerName,
		Header:     fmt.Sprintf("%s: %s", headerName, raw),
		Meta:       fmt.Sprintf(`<meta http-equiv="%s" content="%s">`, headerName, raw),
		Nginx:      fmt.Sprintf(`add_header %s "%s";`, headerName, raw),
		Apache:     fmt.Sprintf(`Header set %s "%s"`, headerName, raw),
	}
}

func isInDirectiveOrder(d string) bool {
	for _, order := range directiveOrder {
		if order == d {
			return true
		}
	}
	return false
}

func parseSourceList(s string) []string {
	parts := strings.Fields(s)
	var res []string
	for _, p := range parts {
		token := normalizeSourceToken(p)
		if token != "" {
			res = append(res, token)
		}
	}
	return res
}

func normalizeSourceToken(token string) string {
	t := strings.TrimSpace(token)
	if t == "" {
		return ""
	}

	lower := strings.ToLower(t)
	if kw, isKw := standardKeywords[lower]; isKw {
		return kw
	}

	// Remove single quotes if wrongly wrapping host or scheme
	if strings.HasPrefix(t, "'") && strings.HasSuffix(t, "'") && len(t) > 2 {
		inner := t[1 : len(t)-1]
		if kw, isKw := standardKeywords[strings.ToLower(inner)]; isKw {
			return kw
		}
		t = inner
	}

	return t
}

func optimizeSources(sources []string) []string {
	if len(sources) == 0 {
		return nil
	}

	// 1. Separate wildcards, exact origins, schemes, keywords
	var wildcards []string
	var nonWildcardHosts []string
	var schemes []string
	var keywords []string
	seen := make(map[string]bool)

	hasSourcesOtherThanNone := false

	for _, s := range sources {
		if seen[s] {
			continue
		}
		seen[s] = true

		if s != "'none'" {
			hasSourcesOtherThanNone = true
		}

		if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
			keywords = append(keywords, s)
		} else if strings.HasSuffix(s, ":") {
			schemes = append(schemes, s)
		} else if strings.Contains(s, "*.") {
			wildcards = append(wildcards, s)
		} else {
			nonWildcardHosts = append(nonWildcardHosts, s)
		}
	}

	// If there are valid sources, remove 'none'
	if hasSourcesOtherThanNone {
		var filteredKeywords []string
		for _, kw := range keywords {
			if kw != "'none'" {
				filteredKeywords = append(filteredKeywords, kw)
			}
		}
		keywords = filteredKeywords
	}

	// Filter out nonWildcardHosts that are covered by any wildcard
	var retainedHosts []string
	for _, host := range nonWildcardHosts {
		covered := false
		for _, wc := range wildcards {
			if isCoveredByWildcard(host, wc) {
				covered = true
				break
			}
		}
		if !covered {
			retainedHosts = append(retainedHosts, host)
		}
	}

	// Sort keywords: 'self' first, then others
	sort.SliceStable(keywords, func(i, j int) bool {
		if keywords[i] == "'self'" {
			return true
		}
		if keywords[j] == "'self'" {
			return false
		}
		return keywords[i] < keywords[j]
	})

	sort.Strings(wildcards)
	sort.Strings(retainedHosts)
	sort.Strings(schemes)

	var result []string
	result = append(result, keywords...)
	result = append(result, wildcards...)
	result = append(result, retainedHosts...)
	result = append(result, schemes...)

	return result
}

// isCoveredByWildcard checks whether origin (e.g., https://api.example.com) is matched by wildcard (https://*.example.com).
func isCoveredByWildcard(origin, wildcard string) bool {
	uOrigin, err1 := url.Parse(origin)
	uWildcard, err2 := url.Parse(wildcard)
	if err1 != nil || err2 != nil {
		return false
	}

	// Schemes must match
	if strings.ToLower(uOrigin.Scheme) != strings.ToLower(uWildcard.Scheme) {
		return false
	}

	// Ports must match (or both default/empty)
	originPort := uOrigin.Port()
	wildcardPort := uWildcard.Port()
	if originPort != wildcardPort {
		return false
	}

	originHost := strings.ToLower(uOrigin.Hostname())
	wildcardHost := strings.ToLower(uWildcard.Hostname())

	if !strings.HasPrefix(wildcardHost, "*.") {
		return false
	}

	domainSuffix := wildcardHost[2:] // e.g. "example.com"
	if originHost == domainSuffix {
		// *.example.com does not cover example.com apex
		return false
	}

	if strings.HasSuffix(originHost, "."+domainSuffix) {
		return true
	}

	return false
}
