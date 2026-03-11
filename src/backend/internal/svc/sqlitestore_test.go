package svc_test

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"open-jarvis/internal/svc"
)

func newTestStore(t *testing.T) *svc.SQLiteConvStore {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() }) //nolint:errcheck // cleanup in test; error logged by sql driver
	store, err := svc.NewSQLiteConvStore(db)
	require.NoError(t, err)
	return store
}

func TestSQLiteGetEmpty(t *testing.T) {
	store := newTestStore(t)
	result := store.Get("missing")
	assert.Nil(t, result)
}

func TestSQLiteSetGet(t *testing.T) {
	store := newTestStore(t)
	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "hello"},
		{Role: openai.ChatMessageRoleAssistant, Content: "world"},
	}
	store.Set("s1", msgs)

	result := store.Get("s1")
	require.Len(t, result, 2)
	assert.Equal(t, openai.ChatMessageRoleUser, result[0].Role)
	assert.Equal(t, "hello", result[0].Content)
	assert.Equal(t, openai.ChatMessageRoleAssistant, result[1].Role)
	assert.Equal(t, "world", result[1].Content)
}

func TestSQLiteSetGetPersists(t *testing.T) {
	// Use a temp file to test persistence across open/close
	tmpFile := t.TempDir() + "/test.db"

	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "persist me"},
	}

	// Write to store
	db1, err := sql.Open("sqlite", tmpFile)
	require.NoError(t, err)
	store1, err := svc.NewSQLiteConvStore(db1)
	require.NoError(t, err)
	store1.Set("s-persist", msgs)
	db1.Close() //nolint:errcheck // cleanup in test; error logged by sql driver

	// Reopen and verify
	db2, err := sql.Open("sqlite", tmpFile)
	require.NoError(t, err)
	t.Cleanup(func() { db2.Close() }) //nolint:errcheck // cleanup in test; error logged by sql driver
	store2, err := svc.NewSQLiteConvStore(db2)
	require.NoError(t, err)

	result := store2.Get("s-persist")
	require.Len(t, result, 1)
	assert.Equal(t, "persist me", result[0].Content)
}

func TestSQLiteList(t *testing.T) {
	store := newTestStore(t)

	err := store.CreateConversation("conv-1", "First Conversation")
	require.NoError(t, err)
	err = store.CreateConversation("conv-2", "Second Conversation")
	require.NoError(t, err)

	convs, err := store.ListConversations()
	require.NoError(t, err)
	require.Len(t, convs, 2)
	// newest first (conv-2 was created last)
	assert.Equal(t, "conv-2", convs[0].ID)
	assert.Equal(t, "conv-1", convs[1].ID)
}

func TestSQLiteDelete(t *testing.T) {
	store := newTestStore(t)

	err := store.CreateConversation("conv-del", "To Delete")
	require.NoError(t, err)

	err = store.DeleteConversation("conv-del")
	require.NoError(t, err)

	convs, err := store.ListConversations()
	require.NoError(t, err)
	assert.Empty(t, convs)
}
