package svc

import (
	"database/sql"
	"fmt"
	"time"
)

const auditSchema = `
CREATE TABLE IF NOT EXISTS tool_audit_log (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp   INTEGER NOT NULL,
    session_id  TEXT NOT NULL DEFAULT '',
    tool_name   TEXT NOT NULL,
    args_json   TEXT NOT NULL DEFAULT '',
    result      TEXT NOT NULL DEFAULT '',
    error       TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_audit_session   ON tool_audit_log(session_id);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON tool_audit_log(timestamp);
`

// AuditStore records every tool execution in an append-only SQLite table.
// Uses the same *sql.DB as SQLiteConvStore — no extra connection needed.
type AuditStore struct {
	db *sql.DB
}

// NewAuditStore creates an AuditStore and runs the schema migration.
func NewAuditStore(db *sql.DB) (*AuditStore, error) {
	a := &AuditStore{db: db}
	if err := a.migrate(); err != nil {
		return nil, fmt.Errorf("audit migrate: %w", err)
	}
	return a, nil
}

func (a *AuditStore) migrate() error {
	_, err := a.db.Exec(auditSchema)
	return err
}

// Log appends one audit record for a tool execution.
// result and errMsg may be empty. Caller is responsible for truncating
// large result strings before passing (recommended: 2000 chars max for audit).
func (a *AuditStore) Log(sessionID, toolName, argsJSON, result, errMsg string) error {
	_, err := a.db.Exec(
		`INSERT INTO tool_audit_log(timestamp, session_id, tool_name, args_json, result, error)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		time.Now().Unix(), sessionID, toolName, argsJSON, result, errMsg,
	)
	return err
}
