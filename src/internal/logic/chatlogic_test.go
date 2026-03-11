package logic_test

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"open-jarvis/internal/config"
	"open-jarvis/internal/logic"
	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// mockStream implements svc.StreamRecver for testing.
type mockStream struct {
	tokens   []string
	pos      int
	err      error // returned after errAfter tokens (or on first call if errAfter=0)
	errAfter int   // -1 means never return err (use io.EOF at end)
}

func (m *mockStream) Recv() (openai.ChatCompletionStreamResponse, error) {
	if m.errAfter >= 0 && m.pos >= m.errAfter {
		if m.err != nil {
			return openai.ChatCompletionStreamResponse{}, m.err
		}
	}
	if m.pos >= len(m.tokens) {
		return openai.ChatCompletionStreamResponse{}, io.EOF
	}
	token := m.tokens[m.pos]
	m.pos++
	return openai.ChatCompletionStreamResponse{
		Choices: []openai.ChatCompletionStreamChoice{
			{Delta: openai.ChatCompletionStreamChoiceDelta{Content: token}},
		},
	}, nil
}

func (m *mockStream) Close() error { return nil }

// mockAIClient implements svc.AIStreamer for testing.
type mockAIClient struct {
	stream  svc.StreamRecver
	err     error
	capture func([]openai.ChatCompletionMessage)
}

func (m *mockAIClient) CreateChatCompletionStream(
	ctx context.Context,
	req openai.ChatCompletionRequest,
) (svc.StreamRecver, error) {
	if m.capture != nil {
		m.capture(req.Messages)
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.stream, nil
}

func newTestSvcCtx(tokens []string, streamErr error, errAfter int, timeoutSecs int) *svc.ServiceContext {
	ms := &mockStream{tokens: tokens, errAfter: errAfter, err: streamErr}
	client := &mockAIClient{stream: ms}
	return svc.NewServiceContextWithClient(config.Config{
		Model: config.ModelConfig{
			SystemPrompt: "You are Jarvis.",
			Name:         "test-model",
		},
		TurnTimeoutSeconds: timeoutSecs,
	}, client, svc.NewConvStore())
}

func TestStreamChatWritesTokens(t *testing.T) {
	svcCtx := newTestSvcCtx([]string{"token1", "token2", "token3"}, nil, -1, 60)
	l := logic.NewChatLogic(context.Background(), svcCtx)

	w := httptest.NewRecorder()
	err := l.StreamChat(&types.ChatRequest{SessionId: "s1", Message: "hi"}, w)
	require.NoError(t, err)

	body := w.Body.String()
	assert.Contains(t, body, "data: token1\n\n")
	assert.Contains(t, body, "data: token2\n\n")
	assert.Contains(t, body, "data: token3\n\n")
}

func TestStreamChatUpdatesHistory(t *testing.T) {
	svcCtx := newTestSvcCtx([]string{"hello", " world"}, nil, -1, 60)
	l := logic.NewChatLogic(context.Background(), svcCtx)

	w := httptest.NewRecorder()
	err := l.StreamChat(&types.ChatRequest{SessionId: "s2", Message: "say hello"}, w)
	require.NoError(t, err)

	history := svcCtx.Store.Get("s2")
	// system + user + assistant
	require.Len(t, history, 3)
	assert.Equal(t, openai.ChatMessageRoleUser, history[1].Role)
	assert.Equal(t, "say hello", history[1].Content)
	assert.Equal(t, openai.ChatMessageRoleAssistant, history[2].Role)
	assert.Equal(t, "hello world", history[2].Content)
}

func TestStreamChatTimeout(t *testing.T) {
	// TurnTimeoutSeconds=0 means immediate timeout; errAfter=0 returns error on first call
	svcCtx := newTestSvcCtx(nil, context.DeadlineExceeded, 0, 0)
	l := logic.NewChatLogic(context.Background(), svcCtx)

	w := httptest.NewRecorder()
	err := l.StreamChat(&types.ChatRequest{SessionId: "s3", Message: "timeout"}, w)
	assert.Error(t, err)
	assert.Contains(t, w.Body.String(), "data: [ERROR]")
}

func TestStreamChatSystemPrompt(t *testing.T) {
	var capturedMessages []openai.ChatCompletionMessage
	ms := &mockStream{tokens: []string{"ok"}, errAfter: -1}
	client := &mockAIClient{
		stream: ms,
		capture: func(msgs []openai.ChatCompletionMessage) {
			capturedMessages = msgs
		},
	}
	svcCtx := svc.NewServiceContextWithClient(config.Config{
		Model: config.ModelConfig{
			SystemPrompt: "You are Jarvis.",
			Name:         "test-model",
		},
		TurnTimeoutSeconds: 60,
	}, client, svc.NewConvStore())

	l := logic.NewChatLogic(context.Background(), svcCtx)
	w := httptest.NewRecorder()
	_ = l.StreamChat(&types.ChatRequest{SessionId: "s4", Message: "hello"}, w)

	require.GreaterOrEqual(t, len(capturedMessages), 2, "expected at least 2 messages (system + user)")
	assert.Equal(t, openai.ChatMessageRoleSystem, capturedMessages[0].Role)
	assert.Equal(t, openai.ChatMessageRoleUser, capturedMessages[1].Role)
}

func TestStreamChatNewSession(t *testing.T) {
	svcCtx := newTestSvcCtx([]string{"hello"}, nil, -1, 60)
	l := logic.NewChatLogic(context.Background(), svcCtx)

	w := httptest.NewRecorder()
	req := &types.ChatRequest{SessionId: "", Message: "tell me something"}
	err := l.StreamChat(req, w)
	require.NoError(t, err)

	// SessionId should be assigned a UUID
	assert.NotEmpty(t, req.SessionId, "expected SessionId to be assigned")

	// Store should have exactly one conversation
	convs, err := svcCtx.Store.ListConversations()
	require.NoError(t, err)
	assert.Len(t, convs, 1)

	// SSE body should contain done event
	body := w.Body.String()
	assert.Contains(t, body, `"done":true`)
}

func TestStreamChatExistingSession(t *testing.T) {
	svcCtx := newTestSvcCtx([]string{"response"}, nil, -1, 60)

	// Pre-seed the store with an existing conversation
	existingID := "existing-id"
	_ = svcCtx.Store.CreateConversation(existingID, "Original Title")
	svcCtx.Store.Set(existingID, []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "previous message"},
	})

	l := logic.NewChatLogic(context.Background(), svcCtx)
	w := httptest.NewRecorder()
	req := &types.ChatRequest{SessionId: existingID, Message: "follow up"}
	err := l.StreamChat(req, w)
	require.NoError(t, err)

	// Session ID should remain unchanged
	assert.Equal(t, existingID, req.SessionId)

	// Conversation should still exist with original title
	conv, err := svcCtx.Store.GetConversation(existingID)
	require.NoError(t, err)
	require.NotNil(t, conv)
	assert.Equal(t, "Original Title", conv.Title)

	// Messages should be appended (previous + new user + assistant)
	msgs := svcCtx.Store.Get(existingID)
	assert.True(t, len(msgs) >= 2, "expected messages to be present")

	// SSE body should contain done event
	body := w.Body.String()
	assert.Contains(t, body, `"done":true`)
}
