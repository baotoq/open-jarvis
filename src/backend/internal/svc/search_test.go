package svc_test

import (
	"database/sql"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"open-jarvis/internal/svc"
)

func newSearchTestStore(t *testing.T) (*svc.SQLiteConvStore, *sql.DB) {
	t.Helper()
	tmpFile := t.TempDir() + "/search_test.db"
	db, err := sql.Open("sqlite", tmpFile)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() }) //nolint:errcheck // cleanup in test; error logged by sql driver
	store, err := svc.NewSQLiteConvStore(db)
	require.NoError(t, err)
	return store, db
}

func TestFTSMigration(t *testing.T) {
	_, db := newSearchTestStore(t)

	// Verify messages_fts virtual table exists in sqlite_master
	var name string
	err := db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name='messages_fts'`,
	).Scan(&name)
	require.NoError(t, err, "messages_fts table should exist after migration")
	assert.Equal(t, "messages_fts", name)

	// Verify FTS5 is functional by running a simple MATCH query (no rows expected)
	rows, err := db.Query(`SELECT rowid FROM messages_fts WHERE messages_fts MATCH ? LIMIT 1`, `"test"`)
	require.NoError(t, err, "FTS5 MATCH query should execute without error")
	rows.Close() //nolint:errcheck // cleanup in test; error logged by sql driver
}

func TestFTSMigration_ExistingRows(t *testing.T) {
	// Simulate existing data: create the base tables and insert a message
	// before calling NewSQLiteConvStore (which runs migrate with initial populate)
	tmpFile := t.TempDir() + "/existing_rows.db"
	db, err := sql.Open("sqlite", tmpFile)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() }) //nolint:errcheck // cleanup in test; error logged by sql driver

	// Create base schema manually (without FTS)
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS conversations (
    id         TEXT PRIMARY KEY,
    title      TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS messages (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role            TEXT NOT NULL,
    content         TEXT NOT NULL,
    position        INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_messages_conv_id ON messages(conversation_id, position);
`)
	require.NoError(t, err)

	// Insert existing data before FTS migration
	_, err = db.Exec(
		`INSERT INTO conversations(id, title, created_at, updated_at) VALUES ('existing-conv', 'Old Chat', 1000, 1000)`,
	)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO messages(conversation_id, role, content, position) VALUES ('existing-conv', 'user', 'hello world existing', 0)`,
	)
	require.NoError(t, err)

	// Now run NewSQLiteConvStore — should run FTS migration + initial populate
	store, err := svc.NewSQLiteConvStore(db)
	require.NoError(t, err)

	// Search should find the pre-existing message
	results, err := store.SearchConversations("hello")
	require.NoError(t, err)
	require.Len(t, results, 1, "should find the pre-existing message via FTS initial populate")
	assert.Equal(t, "existing-conv", results[0].ConversationID)
}

func TestSearchConversations(t *testing.T) {
	store, _ := newSearchTestStore(t)

	// Create conversation and insert a message containing "hello world"
	err := store.CreateConversation("conv-search", "Test Chat")
	require.NoError(t, err)

	// Use Set to insert messages (which triggers FTS sync via INSERT trigger)
	store.Set("conv-search", []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "hello world from user"},
	})

	results, err := store.SearchConversations("hello")
	require.NoError(t, err)
	require.Len(t, results, 1, "should find conversation with matching content")
	assert.Equal(t, "conv-search", results[0].ConversationID)
	assert.NotEmpty(t, results[0].Snippet, "snippet should be non-empty")
}

func TestSearchConversations_NoMatch(t *testing.T) {
	store, _ := newSearchTestStore(t)

	err := store.CreateConversation("conv-nomatch", "Empty Chat")
	require.NoError(t, err)

	results, err := store.SearchConversations("xyznonexistent")
	require.NoError(t, err)
	assert.Empty(t, results, "no match should return empty slice, not error")
}

func TestSearchSanitize(t *testing.T) {
	// sanitizeFTSQuery(`hello "world`) should return `"hello ""world"`
	result := svc.SanitizeFTSQuery(`hello "world`)
	assert.Equal(t, `"hello ""world"`, result)
}

func TestSearchSanitize_Empty(t *testing.T) {
	result := svc.SanitizeFTSQuery("")
	assert.Equal(t, "", result)
}

func TestSearchConversations_SpecialChars(t *testing.T) {
	store, _ := newSearchTestStore(t)

	err := store.CreateConversation("conv-special", "Special Chat")
	require.NoError(t, err)

	// Insert a message with the special char content
	store.Set("conv-special", []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: `hello "world special content`},
	})

	// Search with special FTS5 chars — should not produce error
	results, err := store.SearchConversations(`hello "world`)
	require.NoError(t, err, "special chars in query should not produce error")
	require.Len(t, results, 1, "should find the message with special char content")
	assert.Equal(t, "conv-special", results[0].ConversationID)
}
