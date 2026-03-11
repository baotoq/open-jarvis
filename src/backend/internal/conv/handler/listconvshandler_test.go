package conv_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"open-jarvis/internal/config"
	handler "open-jarvis/internal/conv/handler"
	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// mockStore is a test double for svc.ConversationStore.
type mockStore struct {
	convs []svc.Conversation
	msgs  map[string][]openai.ChatCompletionMessage
}

func newMockStore() *mockStore {
	return &mockStore{
		msgs: make(map[string][]openai.ChatCompletionMessage),
	}
}

func (m *mockStore) Get(sessionID string) []openai.ChatCompletionMessage {
	return m.msgs[sessionID]
}

func (m *mockStore) Set(sessionID string, msgs []openai.ChatCompletionMessage) {
	m.msgs[sessionID] = msgs
}

func (m *mockStore) ListConversations() ([]svc.Conversation, error) {
	return m.convs, nil
}

func (m *mockStore) GetConversation(id string) (*svc.Conversation, error) {
	for i := range m.convs {
		if m.convs[i].ID == id {
			return &m.convs[i], nil
		}
	}
	return nil, nil
}

func (m *mockStore) DeleteConversation(id string) error {
	for i := range m.convs {
		if m.convs[i].ID == id {
			m.convs = append(m.convs[:i], m.convs[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockStore) CreateConversation(id, title string) error {
	m.convs = append(m.convs, svc.Conversation{ID: id, Title: title})
	return nil
}

func newTestHandlerSvcCtx(store svc.ConversationStore) *svc.ServiceContext {
	return svc.NewServiceContextWithClient(config.Config{}, nil, store)
}

func TestListConversations(t *testing.T) {
	store := newMockStore()
	_ = store.CreateConversation("conv-1", "First conversation")
	_ = store.CreateConversation("conv-2", "Second conversation")

	svcCtx := newTestHandlerSvcCtx(store)
	h := handler.ListConversationsHandler(svcCtx)

	r := httptest.NewRequest(http.MethodGet, "/api/conversations", nil)
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []types.ConversationResponse
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}
