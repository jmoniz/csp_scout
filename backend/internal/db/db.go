package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"csp_report_backend/internal/model"

	_ "modernc.org/sqlite"
)

// DB wraps the SQL database handle.
type DB struct {
	*sql.DB
}

// Open initializes SQLite connection with WAL mode and pragmas.
func Open(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Apply critical SQLite pragmas
	pragmas := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
	}

	// WAL mode is only valid for file-based databases, not pure in-memory
	if !strings.Contains(dataSourceName, ":memory:") {
		pragmas = append(pragmas, "PRAGMA journal_mode = WAL;", "PRAGMA synchronous = NORMAL;")
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to execute pragma '%s': %w", pragma, err)
		}
	}

	wrapper := &DB{db}
	if err := wrapper.Migrate(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return wrapper, nil
}

// Migrate creates tables and indexes if they don't exist.
func (d *DB) Migrate(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS session (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		target_origin TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS raw_report (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		document_uri TEXT,
		referrer TEXT,
		violated_directive TEXT NOT NULL,
		effective_directive TEXT NOT NULL,
		original_policy TEXT,
		disposition TEXT,
		blocked_uri TEXT NOT NULL,
		line_number INTEGER,
		column_number INTEGER,
		source_file TEXT,
		status_code INTEGER,
		script_sample TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(session_id) REFERENCES session(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS violation_group (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		directive TEXT NOT NULL,
		origin_host TEXT NOT NULL,
		suggested_wildcard TEXT,
		is_self BOOLEAN DEFAULT 0,
		count INTEGER DEFAULT 1,
		last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		status TEXT DEFAULT 'pending',
		UNIQUE(session_id, directive, origin_host),
		FOREIGN KEY(session_id) REFERENCES session(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS session_policy_setting (
		session_id TEXT PRIMARY KEY,
		default_src TEXT DEFAULT "'none'",
		form_action TEXT DEFAULT "'none'",
		frame_ancestors TEXT DEFAULT "'none'",
		upgrade_insecure_requests BOOLEAN DEFAULT 0,
		block_all_mixed_content BOOLEAN DEFAULT 0,
		report_only BOOLEAN DEFAULT 0,
		custom_directives TEXT DEFAULT '',
		FOREIGN KEY(session_id) REFERENCES session(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_raw_report_session_id ON raw_report(session_id);
	CREATE INDEX IF NOT EXISTS idx_violation_group_lookup ON violation_group(session_id, directive, status);
	`

	_, err := d.ExecContext(ctx, schema)
	return err
}

// CreateSession inserts a new session and its default policy setting.
func (d *DB) CreateSession(ctx context.Context, s *model.Session) error {
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now

	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO session (id, name, target_origin, description, created_at, updated_at) 
	          VALUES (?, ?, ?, ?, ?, ?)`
	if _, err := tx.ExecContext(ctx, query, s.ID, s.Name, s.TargetOrigin, s.Description, s.CreatedAt, s.UpdatedAt); err != nil {
		return fmt.Errorf("insert session failed: %w", err)
	}

	settingQuery := `INSERT INTO session_policy_setting (session_id, default_src, form_action, frame_ancestors)
	                VALUES (?, "'none'", "'none'", "'none'")`
	if _, err := tx.ExecContext(ctx, settingQuery, s.ID); err != nil {
		return fmt.Errorf("insert default policy setting failed: %w", err)
	}

	return tx.Commit()
}

// GetSession retrieves a session by ID with aggregate counts.
func (d *DB) GetSession(ctx context.Context, id string) (*model.Session, error) {
	query := `
	SELECT s.id, s.name, s.target_origin, COALESCE(s.description, ''), s.created_at, s.updated_at,
	       (SELECT COUNT(*) FROM raw_report WHERE session_id = s.id) AS total_reports,
	       (SELECT COUNT(*) FROM violation_group WHERE session_id = s.id AND status = 'pending') AS pending_violations,
	       (SELECT COUNT(*) FROM violation_group WHERE session_id = s.id AND status LIKE 'approved%') AS approved_rules,
	       (SELECT COUNT(*) FROM violation_group WHERE session_id = s.id AND is_self = 1) AS self_violations
	FROM session s
	WHERE s.id = ?`

	var s model.Session
	var createdAtStr, updatedAtStr string
	err := d.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.TargetOrigin, &s.Description, &createdAtStr, &updatedAtStr,
		&s.TotalReports, &s.PendingViolations, &s.ApprovedRules, &s.SelfViolations,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	s.CreatedAt = parseTime(createdAtStr)
	s.UpdatedAt = parseTime(updatedAtStr)
	return &s, nil
}

// ListSessions retrieves all sessions ordered by creation date descending.
func (d *DB) ListSessions(ctx context.Context) ([]model.Session, error) {
	query := `
	SELECT s.id, s.name, s.target_origin, COALESCE(s.description, ''), s.created_at, s.updated_at,
	       (SELECT COUNT(*) FROM raw_report WHERE session_id = s.id) AS total_reports,
	       (SELECT COUNT(*) FROM violation_group WHERE session_id = s.id AND status = 'pending') AS pending_violations,
	       (SELECT COUNT(*) FROM violation_group WHERE session_id = s.id AND status LIKE 'approved%') AS approved_rules,
	       (SELECT COUNT(*) FROM violation_group WHERE session_id = s.id AND is_self = 1) AS self_violations
	FROM session s
	ORDER BY s.created_at DESC`

	rows, err := d.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []model.Session
	for rows.Next() {
		var s model.Session
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(&s.ID, &s.Name, &s.TargetOrigin, &s.Description, &createdAtStr, &updatedAtStr,
			&s.TotalReports, &s.PendingViolations, &s.ApprovedRules, &s.SelfViolations); err != nil {
			return nil, err
		}
		s.CreatedAt = parseTime(createdAtStr)
		s.UpdatedAt = parseTime(updatedAtStr)
		sessions = append(sessions, s)
	}

	if sessions == nil {
		sessions = []model.Session{}
	}
	return sessions, rows.Err()
}

// DeleteSession deletes a session and cascades related rows.
func (d *DB) DeleteSession(ctx context.Context, id string) (bool, error) {
	res, err := d.ExecContext(ctx, `DELETE FROM session WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

// IngestReport logs the raw report and upserts the aggregated violation_group.
func (d *DB) IngestReport(ctx context.Context, nv *model.NormalizedViolation, sessionID string) error {
	now := time.Now().UTC()

	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert raw report
	rawQuery := `INSERT INTO raw_report (
		session_id, document_uri, referrer, violated_directive, effective_directive,
		original_policy, disposition, blocked_uri, line_number, column_number,
		source_file, status_code, script_sample, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, rawQuery,
		sessionID, nv.DocumentURI, nv.Referrer, nv.ViolatedDirective, nv.EffectiveDirective,
		nv.OriginalPolicy, nv.Disposition, nv.BlockedURI, nv.LineNumber, nv.ColumnNumber,
		nv.SourceFile, nv.StatusCode, nv.ScriptSample, now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert raw report: %w", err)
	}

	// 2. Upsert violation group
	upsertQuery := `
	INSERT INTO violation_group (session_id, directive, origin_host, suggested_wildcard, is_self, count, last_seen_at, status)
	VALUES (?, ?, ?, ?, ?, 1, ?, 'pending')
	ON CONFLICT(session_id, directive, origin_host) DO UPDATE SET
		count = count + 1,
		last_seen_at = excluded.last_seen_at;
	`
	_, err = tx.ExecContext(ctx, upsertQuery,
		sessionID, nv.EffectiveDirective, nv.OriginHost, nv.SuggestedWildcard, nv.IsSelf, now,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert violation group: %w", err)
	}

	return tx.Commit()
}

// ListViolationGroups returns grouped violations filtered by directive, status, search string, or source type.
func (d *DB) ListViolationGroups(ctx context.Context, sessionID, directive, status, search, sourceType string) ([]model.ViolationGroup, error) {
	var conditions []string
	var args []any

	conditions = append(conditions, "session_id = ?")
	args = append(args, sessionID)

	if directive != "" && directive != "all" {
		conditions = append(conditions, "directive = ?")
		args = append(args, directive)
	}
	if status != "" && status != "all" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}
	if search != "" {
		conditions = append(conditions, "(origin_host LIKE ? OR suggested_wildcard LIKE ?)")
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	if sourceType == "self" {
		conditions = append(conditions, "is_self = 1")
	} else if sourceType == "third_party" {
		conditions = append(conditions, "is_self = 0")
	}

	query := fmt.Sprintf(`
	SELECT id, session_id, directive, origin_host, COALESCE(suggested_wildcard, ''), is_self, count, last_seen_at, status
	FROM violation_group
	WHERE %s
	ORDER BY count DESC, last_seen_at DESC`, strings.Join(conditions, " AND "))

	rows, err := d.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []model.ViolationGroup
	for rows.Next() {
		var vg model.ViolationGroup
		var lastSeenStr string
		if err := rows.Scan(&vg.ID, &vg.SessionID, &vg.Directive, &vg.OriginHost, &vg.SuggestedWildcard,
			&vg.IsSelf, &vg.Count, &lastSeenStr, &vg.Status); err != nil {
			return nil, err
		}
		vg.LastSeenAt = parseTime(lastSeenStr)
		groups = append(groups, vg)
	}

	if groups == nil {
		groups = []model.ViolationGroup{}
	}
	return groups, rows.Err()
}

// GetViolationGroup fetches a single violation group by ID.
func (d *DB) GetViolationGroup(ctx context.Context, id int64) (*model.ViolationGroup, error) {
	query := `SELECT id, session_id, directive, origin_host, COALESCE(suggested_wildcard, ''), is_self, count, last_seen_at, status
	          FROM violation_group WHERE id = ?`

	var vg model.ViolationGroup
	var lastSeenStr string
	err := d.QueryRowContext(ctx, query, id).Scan(
		&vg.ID, &vg.SessionID, &vg.Directive, &vg.OriginHost, &vg.SuggestedWildcard,
		&vg.IsSelf, &vg.Count, &lastSeenStr, &vg.Status,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	vg.LastSeenAt = parseTime(lastSeenStr)
	return &vg, nil
}

// UpdateViolationGroupStatus updates the decision status for a single violation.
func (d *DB) UpdateViolationGroupStatus(ctx context.Context, id int64, status string) (bool, error) {
	res, err := d.ExecContext(ctx, `UPDATE violation_group SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return false, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

// BulkUpdateViolationGroupStatus updates the decision status for multiple violations in a session.
func (d *DB) BulkUpdateViolationGroupStatus(ctx context.Context, sessionID string, ids []int64, status string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+2)
	args = append(args, status, sessionID)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(`UPDATE violation_group SET status = ? WHERE session_id = ? AND id IN (%s)`, strings.Join(placeholders, ","))
	res, err := d.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// BulkApproveSelf marks all pending 'self' violations in a session as 'approved_self'.
func (d *DB) BulkApproveSelf(ctx context.Context, sessionID string) (int64, error) {
	res, err := d.ExecContext(ctx, `UPDATE violation_group SET status = 'approved_self' WHERE session_id = ? AND is_self = 1 AND status = 'pending'`, sessionID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// GetRawReportSamples returns up to limit recent raw reports matching a session and origin host.
func (d *DB) GetRawReportSamples(ctx context.Context, sessionID, directive, originHost string, limit int) ([]model.RawReport, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `
	SELECT id, session_id, COALESCE(document_uri, ''), COALESCE(referrer, ''), violated_directive, effective_directive,
	       COALESCE(original_policy, ''), COALESCE(disposition, ''), blocked_uri, COALESCE(line_number, 0),
	       COALESCE(column_number, 0), COALESCE(source_file, ''), COALESCE(status_code, 0), COALESCE(script_sample, ''),
	       created_at
	FROM raw_report
	WHERE session_id = ? AND effective_directive = ? AND blocked_uri LIKE ?
	ORDER BY id DESC
	LIMIT ?`

	rows, err := d.QueryContext(ctx, query, sessionID, directive, "%"+originHost+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []model.RawReport
	for rows.Next() {
		var r model.RawReport
		var createdAtStr string
		if err := rows.Scan(
			&r.ID, &r.SessionID, &r.DocumentURI, &r.Referrer, &r.ViolatedDirective, &r.EffectiveDirective,
			&r.OriginalPolicy, &r.Disposition, &r.BlockedURI, &r.LineNumber, &r.ColumnNumber,
			&r.SourceFile, &r.StatusCode, &r.ScriptSample, &createdAtStr,
		); err != nil {
			return nil, err
		}
		r.CreatedAt = parseTime(createdAtStr)
		reports = append(reports, r)
	}
	if reports == nil {
		reports = []model.RawReport{}
	}
	return reports, rows.Err()
}

// GetSessionPolicySetting retrieves custom policy settings for a session.
func (d *DB) GetSessionPolicySetting(ctx context.Context, sessionID string) (*model.SessionPolicySetting, error) {
	query := `SELECT session_id, default_src, form_action, frame_ancestors, upgrade_insecure_requests,
	                 block_all_mixed_content, report_only, COALESCE(custom_directives, '')
	          FROM session_policy_setting WHERE session_id = ?`

	var ps model.SessionPolicySetting
	err := d.QueryRowContext(ctx, query, sessionID).Scan(
		&ps.SessionID, &ps.DefaultSrc, &ps.FormAction, &ps.FrameAncestors,
		&ps.UpgradeInsecureRequests, &ps.BlockAllMixedContent, &ps.ReportOnly, &ps.CustomDirectives,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ps, nil
}

// UpdateSessionPolicySetting updates custom policy settings for a session.
func (d *DB) UpdateSessionPolicySetting(ctx context.Context, ps *model.SessionPolicySetting) error {
	query := `
	INSERT INTO session_policy_setting (session_id, default_src, form_action, frame_ancestors,
	                                   upgrade_insecure_requests, block_all_mixed_content, report_only, custom_directives)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(session_id) DO UPDATE SET
		default_src = excluded.default_src,
		form_action = excluded.form_action,
		frame_ancestors = excluded.frame_ancestors,
		upgrade_insecure_requests = excluded.upgrade_insecure_requests,
		block_all_mixed_content = excluded.block_all_mixed_content,
		report_only = excluded.report_only,
		custom_directives = excluded.custom_directives;`

	_, err := d.ExecContext(ctx, query,
		ps.SessionID, ps.DefaultSrc, ps.FormAction, ps.FrameAncestors,
		ps.UpgradeInsecureRequests, ps.BlockAllMixedContent, ps.ReportOnly, ps.CustomDirectives,
	)
	return err
}

func parseTime(val string) time.Time {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, val); err == nil {
			return t
		}
	}
	return time.Time{}
}
