package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"open-jarvis/internal/handler"
	"open-jarvis/internal/types"
)

func TestGetConversationMessages_found(t *testing.T) {
	store := newMockStore()
	store.msgs["s1"] = []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "hello"},
		{Role: openai.ChatMessageRoleAssistant, Content: "hi there"},
	}

	svcCtx := newTestHandlerSvcCtx(store)
	h := handler.GetConversationMessagesHandler(svcCtx)

	r := httptest.NewRequest(http.MethodGet, "/api/conversations/s1/messages", nil)
	r = r.WithContext(context.Background())
	r = injectPathParam(r, "id", "s1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []types.MessageResponse
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "user", result[0].Role)
	assert.Equal(t, "hello", result[0].Content)
	assert.Equal(t, "assistant", result[1].Role)
	assert.Equal(t, "hi there", result[1].Content)
}

func TestGetConversationMessages_empty(t *testing.T) {
	store := newMockStore()

	svcCtx := newTestHandlerSvcCtx(store)
	h := handler.GetConversationMessagesHandler(svcCtx)

	r := httptest.NewRequest(http.MethodGet, "/api/conversations/s2/messages", nil)
	r = r.WithContext(context.Background())
	r = injectPathParam(r, "id", "s2")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.JSONEq(t, "[]", body)
}
