package svc_test

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"open-jarvis/internal/svc"
)

func newTestAuditStore(t *testing.T) (*svc.AuditStore, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	store, err := svc.NewAuditStore(db)
	require.NoError(t, err)
	return store, db
}

func TestAuditStore(t *testing.T) {
	t.Run("migrate idempotent", func(t *testing.T) {
		db, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		t.Cleanup(func() { db.Close() })

		// First creation
		_, err = svc.NewAuditStore(db)
		require.NoError(t, err)

		// Second creation on same DB must also succeed (idempotent)
		_, err = svc.NewAuditStore(db)
		require.NoError(t, err)
	})

	t.Run("log inserts row", func(t *testing.T) {
		store, db := newTestAuditStore(t)

		err := store.Log("session-1", "web_search", `{"query":"go test"}`, "results here", "")
		require.NoError(t, err)

		var ts int64
		var sessionID, toolName, argsJSON, result, errMsg string
		err = db.QueryRow(
			`SELECT timestamp, session_id, tool_name, args_json, result, error FROM tool_audit_log LIMIT 1`,
		).Scan(&ts, &sessionID, &toolName, &argsJSON, &result, &errMsg)
		require.NoError(t, err)

		assert.Greater(t, ts, int64(0))
		assert.Equal(t, "session-1", sessionID)
		assert.Equal(t, "web_search", toolName)
		assert.Equal(t, `{"query":"go test"}`, argsJSON)
		assert.Equal(t, "results here", result)
		assert.Equal(t, "", errMsg)
	})

	t.Run("log multiple rows", func(t *testing.T) {
		store, db := newTestAuditStore(t)

		require.NoError(t, store.Log("s1", "tool_a", "{}", "r1", ""))
		require.NoError(t, store.Log("s1", "tool_b", "{}", "r2", "err1"))
		require.NoError(t, store.Log("s2", "tool_c", `{"x":1}`, "r3", ""))

		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM tool_audit_log`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("empty strings stored", func(t *testing.T) {
		store, db := newTestAuditStore(t)

		err := store.Log("", "tool_x", "{}", "", "")
		require.NoError(t, err)

		var sessionID, result, errMsg string
		err = db.QueryRow(
			`SELECT session_id, result, error FROM tool_audit_log LIMIT 1`,
		).Scan(&sessionID, &result, &errMsg)
		require.NoError(t, err)

		assert.Equal(t, "", sessionID)
		assert.Equal(t, "", result)
		assert.Equal(t, "", errMsg)
	})
}
