package db

import (
	"context"
	"testing"

	"csp_report_backend/internal/model"
)

func TestDB_Lifecycle(t *testing.T) {
	ctx := context.Background()
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer database.Close()

	// 1. Create Session
	session := &model.Session{
		ID:           "test-session-uuid",
		Name:         "Portal Test",
		TargetOrigin: "https://app.example.com",
		Description:  "Testing CSP",
	}
	if err := database.CreateSession(ctx, session); err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// 2. Get Session
	fetched, err := database.GetSession(ctx, "test-session-uuid")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if fetched == nil || fetched.Name != "Portal Test" {
		t.Fatalf("unexpected fetched session: %+v", fetched)
	}

	// 3. Ingest Violation
	nv := &model.NormalizedViolation{
		DocumentURI:        "https://app.example.com/dashboard",
		Referrer:           "https://app.example.com/",
		ViolatedDirective:  "connect-src",
		EffectiveDirective: "connect-src",
		OriginalPolicy:     "default-src 'none'",
		Disposition:        "report",
		BlockedURI:         "https://api.example.com/v1/resources",
		LineNumber:         1,
		ColumnNumber:       25897,
		SourceFile:         "https://app.example.com/bundle.js",
		StatusCode:         200,
		ScriptSample:       "",
		OriginHost:         "https://api.example.com",
		SuggestedWildcard:  "https://*.example.com",
		IsSelf:             false,
	}

	if err := database.IngestReport(ctx, nv, session.ID); err != nil {
		t.Fatalf("IngestReport failed: %v", err)
	}

	// Duplicate ingest should increment count
	if err := database.IngestReport(ctx, nv, session.ID); err != nil {
		t.Fatalf("IngestReport duplicate failed: %v", err)
	}

	// 4. List Violation Groups
	groups, err := database.ListViolationGroups(ctx, session.ID, "connect-src", "", "", "")
	if err != nil {
		t.Fatalf("ListViolationGroups failed: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Count != 2 {
		t.Fatalf("expected count 2, got %d", groups[0].Count)
	}
	if groups[0].SuggestedWildcard != "https://*.example.com" {
		t.Fatalf("expected wildcard https://*.example.com, got %s", groups[0].SuggestedWildcard)
	}

	// 5. Update Status
	updated, err := database.UpdateViolationGroupStatus(ctx, groups[0].ID, "approved_origin")
	if err != nil || !updated {
		t.Fatalf("UpdateViolationGroupStatus failed: %v", err)
	}

	// 6. Check updated session stats
	fetched, err = database.GetSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if fetched.TotalReports != 2 {
		t.Fatalf("expected 2 total reports, got %d", fetched.TotalReports)
	}
	if fetched.ApprovedRules != 1 {
		t.Fatalf("expected 1 approved rule, got %d", fetched.ApprovedRules)
	}
	if fetched.PendingViolations != 0 {
		t.Fatalf("expected 0 pending violations, got %d", fetched.PendingViolations)
	}

	// 7. Policy Settings
	ps, err := database.GetSessionPolicySetting(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetSessionPolicySetting failed: %v", err)
	}
	if ps.DefaultSrc != "'none'" {
		t.Fatalf("expected 'none', got %s", ps.DefaultSrc)
	}
	ps.UpgradeInsecureRequests = true
	if err := database.UpdateSessionPolicySetting(ctx, ps); err != nil {
		t.Fatalf("UpdateSessionPolicySetting failed: %v", err)
	}
	ps, err = database.GetSessionPolicySetting(ctx, session.ID)
	if err != nil || !ps.UpgradeInsecureRequests {
		t.Fatalf("expected UpgradeInsecureRequests true, got %+v", ps)
	}

	// 8. Delete session
	deleted, err := database.DeleteSession(ctx, session.ID)
	if err != nil || !deleted {
		t.Fatalf("DeleteSession failed: %v", err)
	}
	groups, err = database.ListViolationGroups(ctx, session.ID, "", "", "", "")
	if err != nil {
		t.Fatalf("ListViolationGroups after delete failed: %v", err)
	}
	if len(groups) != 0 {
		t.Fatalf("expected 0 groups after cascade delete, got %d", len(groups))
	}
}
