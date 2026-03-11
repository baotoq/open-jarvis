package logic_test

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	openai "github.com/sashabaranov/go-openai"
	"open-jarvis/internal/config"
	"open-jarvis/internal/logic"
	"open-jarvis/internal/svc"
)

// newInMemorySQLiteStore creates a real SQLiteConvStore backed by an in-memory SQLite DB.
func newInMemorySQLiteStore(t *testing.T) *svc.SQLiteConvStore {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() }) //nolint:errcheck // cleanup in test; error logged by sql driver
	store, err := svc.NewSQLiteConvStore(db)
	require.NoError(t, err)
	return store
}

func TestSearchConvsLogic(t *testing.T) {
	store := newInMemorySQLiteStore(t)

	// Insert a conversation with a searchable message.
	convID := "conv-search-1"
	err := store.CreateConversation(convID, "Test Conversation")
	require.NoError(t, err)
	store.Set(convID, []openai.ChatCompletionMessage{
		{Role: "user", Content: "hello world keyword"},
	})

	svcCtx := svc.NewServiceContextForTest(config.Config{}, &mockAIClient{}, store)
	l := logic.NewSearchConvsLogic(context.Background(), svcCtx)

	results, err := l.Search("keyword")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, convID, results[0].ID)
}

func TestSearchConvsLogic_Empty(t *testing.T) {
	store := newInMemorySQLiteStore(t)
	svcCtx := svc.NewServiceContextForTest(config.Config{}, &mockAIClient{}, store)
	l := logic.NewSearchConvsLogic(context.Background(), svcCtx)

	results, err := l.Search("")
	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchConvsLogic_NoStore(t *testing.T) {
	// Use an in-memory ConvStore (mockConvStore) that does NOT implement ConvSearcher.
	svcCtx := svc.NewServiceContextForTest(config.Config{}, &mockAIClient{}, newMockConvStore())
	l := logic.NewSearchConvsLogic(context.Background(), svcCtx)

	results, err := l.Search("anything")
	assert.NoError(t, err)
	assert.Nil(t, results)
}
