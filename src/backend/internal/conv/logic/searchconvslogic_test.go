package conv_test

import (
	"context"
	"database/sql"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"open-jarvis/internal/config"
	"open-jarvis/internal/conv/logic"
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

// mockAIClient implements svc.AIStreamer for testing (minimal stub).
type mockAIClient struct{}

func (m *mockAIClient) CreateChatCompletionStream(
	_ context.Context,
	_ openai.ChatCompletionRequest,
) (svc.StreamRecver, error) {
	return nil, nil
}

// mockConvStore is a minimal in-memory ConversationStore for tests.
type mockConvStore struct {
	msgs  map[string][]openai.ChatCompletionMessage
	convs map[string]svc.Conversation
}

func newMockConvStore() *mockConvStore {
	return &mockConvStore{
		msgs:  make(map[string][]openai.ChatCompletionMessage),
		convs: make(map[string]svc.Conversation),
	}
}

func (m *mockConvStore) Get(id string) []openai.ChatCompletionMessage { return m.msgs[id] }
func (m *mockConvStore) Set(id string, msgs []openai.ChatCompletionMessage) { m.msgs[id] = msgs }
func (m *mockConvStore) ListConversations() ([]svc.Conversation, error) {
	result := make([]svc.Conversation, 0, len(m.convs))
	for _, c := range m.convs {
		result = append(result, c)
	}
	return result, nil
}
func (m *mockConvStore) GetConversation(id string) (*svc.Conversation, error) {
	c, ok := m.convs[id]
	if !ok {
		return nil, nil
	}
	return &c, nil
}
func (m *mockConvStore) DeleteConversation(id string) error {
	delete(m.convs, id)
	delete(m.msgs, id)
	return nil
}
func (m *mockConvStore) CreateConversation(id, title string) error {
	m.convs[id] = svc.Conversation{ID: id, Title: title}
	return nil
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
