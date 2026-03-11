package conv_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"open-jarvis/internal/config"
	handler "open-jarvis/internal/conv/handler"
	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

func newSQLiteStoreForTest(t *testing.T) *svc.SQLiteConvStore {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	store, err := svc.NewSQLiteConvStore(db)
	require.NoError(t, err)
	return store
}

func TestSearchConvsHandler(t *testing.T) {
	store := newSQLiteStoreForTest(t)
	convID := "conv-abc"
	require.NoError(t, store.CreateConversation(convID, "My Conversation"))
	store.Set(convID, []openai.ChatCompletionMessage{
		{Role: "user", Content: "hello searchterm world"},
	})

	svcCtx := svc.NewServiceContextForTest(config.Config{}, nil, store)
	h := handler.SearchConversationsHandler(svcCtx)

	r := httptest.NewRequest(http.MethodGet, "/api/conversations/search?q=searchterm", nil)
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var results []types.SearchResult
	err := json.NewDecoder(w.Body).Decode(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, convID, results[0].ID)
}

func TestSearchConvsHandler_EmptyQuery(t *testing.T) {
	store := newSQLiteStoreForTest(t)
	svcCtx := svc.NewServiceContextForTest(config.Config{}, nil, store)
	h := handler.SearchConversationsHandler(svcCtx)

	r := httptest.NewRequest(http.MethodGet, "/api/conversations/search?q=", nil)
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var results []types.SearchResult
	err := json.NewDecoder(w.Body).Decode(&results)
	require.NoError(t, err)
	assert.Empty(t, results)
}
