package chat_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	logic "open-jarvis/internal/chat/logic"
	"open-jarvis/internal/config"
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

// toolCallStream simulates a stream that first emits a tool_calls finish reason,
// then (on second call) emits normal text with stop.
type toolCallStream struct {
	responses []openai.ChatCompletionStreamResponse
	pos       int
}

func (t *toolCallStream) Recv() (openai.ChatCompletionStreamResponse, error) {
	if t.pos >= len(t.responses) {
		return openai.ChatCompletionStreamResponse{}, io.EOF
	}
	r := t.responses[t.pos]
	t.pos++
	return r, nil
}

func (t *toolCallStream) Close() error { return nil }

// mockAIClient implements svc.AIStreamer for testing.
type mockAIClient struct {
	stream  svc.StreamRecver
	err     error
	capture func([]openai.ChatCompletionMessage)
}

func (m *mockAIClient) CreateChatCompletionStream(
	_ context.Context,
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

// multiStreamAIClient returns a different stream on each call.
type multiStreamAIClient struct {
	streams  []svc.StreamRecver
	callIdx  int
	captures [][]openai.ChatCompletionMessage
}

func (m *multiStreamAIClient) CreateChatCompletionStream(
	_ context.Context,
	req openai.ChatCompletionRequest,
) (svc.StreamRecver, error) {
	m.captures = append(m.captures, req.Messages)
	if m.callIdx >= len(m.streams) {
		return nil, io.EOF
	}
	s := m.streams[m.callIdx]
	m.callIdx++
	return s, nil
}

// mockConvStore is a fully functional in-memory ConversationStore for tests
// that need ListConversations / GetConversation / CreateConversation to work.
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

func (m *mockConvStore) Get(id string) []openai.ChatCompletionMessage {
	return m.msgs[id]
}

func (m *mockConvStore) Set(id string, msgs []openai.ChatCompletionMessage) {
	m.msgs[id] = msgs
}

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

func newTestSvcCtx(tokens []string, streamErr error, errAfter int, timeoutSecs int) *svc.ServiceContext {
	ms := &mockStream{tokens: tokens, errAfter: errAfter, err: streamErr}
	client := &mockAIClient{stream: ms}
	return svc.NewServiceContextWithClient(config.Config{
		Model: config.ModelConfig{
			SystemPrompt: "You are Jarvis.",
			Name:         "test-model",
		},
		TurnTimeoutSeconds: timeoutSecs,
	}, client, newMockConvStore())
}

func TestStreamChatWritesTokens(t *testing.T) {
	svcCtx := newTestSvcCtx([]string{"token1", "token2", "token3"}, nil, -1, 60)
	l := logic.NewChatLogic(context.Background(), svcCtx)

	w := httptest.NewRecorder()
	err := l.StreamChat(&types.ChatRequest{SessionID: "s1", Message: "hi"}, w)
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
	// Use a pointer so we can read the assigned session ID after StreamChat.
	// When no prior messages exist for an ID, a new UUID is assigned.
	req := &types.ChatRequest{SessionID: "", Message: "say hello"}
	err := l.StreamChat(req, w)
	require.NoError(t, err)
	require.NotEmpty(t, req.SessionID)

	history := svcCtx.Store.Get(req.SessionID)
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
	err := l.StreamChat(&types.ChatRequest{SessionID: "s3", Message: "timeout"}, w)
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
	}, client, newMockConvStore())

	l := logic.NewChatLogic(context.Background(), svcCtx)
	w := httptest.NewRecorder()
	_ = l.StreamChat(&types.ChatRequest{SessionID: "s4", Message: "hello"}, w)

	require.GreaterOrEqual(t, len(capturedMessages), 2, "expected at least 2 messages (system + user)")
	assert.Equal(t, openai.ChatMessageRoleSystem, capturedMessages[0].Role)
	assert.Equal(t, openai.ChatMessageRoleUser, capturedMessages[1].Role)
}

func TestStreamChatNewSession(t *testing.T) {
	svcCtx := newTestSvcCtx([]string{"hello"}, nil, -1, 60)
	l := logic.NewChatLogic(context.Background(), svcCtx)

	w := httptest.NewRecorder()
	req := &types.ChatRequest{SessionID: "", Message: "tell me something"}
	err := l.StreamChat(req, w)
	require.NoError(t, err)

	// SessionID should be assigned a UUID
	assert.NotEmpty(t, req.SessionID, "expected SessionID to be assigned")

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
	req := &types.ChatRequest{SessionID: existingID, Message: "follow up"}
	err := l.StreamChat(req, w)
	require.NoError(t, err)

	// Session ID should remain unchanged
	assert.Equal(t, existingID, req.SessionID)

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

// makeToolCallsStream builds a stream that emits a single response with FinishReasonToolCalls
// and a read_file tool call accumulated across two delta chunks.
func makeToolCallsStream(toolCallID, funcName, argsJSON string) svc.StreamRecver {
	idx := 0
	// First chunk: tool call header
	tc1 := openai.ToolCall{
		Index: &idx,
		ID:    toolCallID,
		Type:  openai.ToolTypeFunction,
		Function: openai.FunctionCall{
			Name:      funcName,
			Arguments: "",
		},
	}
	// Second chunk: arguments fragment
	tc2 := openai.ToolCall{
		Index: &idx,
		Function: openai.FunctionCall{
			Arguments: argsJSON,
		},
	}
	return &toolCallStream{
		responses: []openai.ChatCompletionStreamResponse{
			{
				Choices: []openai.ChatCompletionStreamChoice{
					{
						Delta: openai.ChatCompletionStreamChoiceDelta{
							ToolCalls: []openai.ToolCall{tc1},
						},
					},
				},
			},
			{
				Choices: []openai.ChatCompletionStreamChoice{
					{
						Delta: openai.ChatCompletionStreamChoiceDelta{
							ToolCalls: []openai.ToolCall{tc2},
						},
					},
				},
			},
			{
				Choices: []openai.ChatCompletionStreamChoice{
					{
						FinishReason: openai.FinishReasonToolCalls,
					},
				},
			},
		},
	}
}

// makeStopStream builds a simple stream that emits text tokens then stops.
func makeStopStream(tokens ...string) svc.StreamRecver {
	responses := make([]openai.ChatCompletionStreamResponse, 0, len(tokens)+1)
	for _, t := range tokens {
		responses = append(responses, openai.ChatCompletionStreamResponse{
			Choices: []openai.ChatCompletionStreamChoice{
				{Delta: openai.ChatCompletionStreamChoiceDelta{Content: t}},
			},
		})
	}
	responses = append(responses, openai.ChatCompletionStreamResponse{
		Choices: []openai.ChatCompletionStreamChoice{
			{FinishReason: openai.FinishReasonStop},
		},
	})
	return &toolCallStream{responses: responses}
}

// TestChatLogic_ToolLoop: mock AIStreamer returns tool_calls finish reason + read_file call,
// then stop finish reason. Verifies tool_call and tool_result SSE events are emitted,
// and final text is streamed.
func TestChatLogic_ToolLoop(t *testing.T) {
	toolCallID := "call_abc123"
	argsJSON := `{"path":"foo.txt"}`

	firstStream := makeToolCallsStream(toolCallID, "read_file", argsJSON)
	secondStream := makeStopStream("file processed")

	multiClient := &multiStreamAIClient{
		streams: []svc.StreamRecver{firstStream, secondStream},
	}

	svcCtx := svc.NewServiceContextForTest(config.Config{
		Model: config.ModelConfig{
			SystemPrompt: "You are Jarvis.",
			Name:         "test-model",
		},
		TurnTimeoutSeconds: 60,
		MaxToolCalls:       5,
		WorkspaceRoot:      t.TempDir(),
	}, multiClient, newMockConvStore())

	// Write a test file so read_file can succeed
	testFile := svcCtx.Config.WorkspaceRoot + "/foo.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("hello from file"), 0644))

	l := logic.NewChatLogic(context.Background(), svcCtx)
	w := httptest.NewRecorder()
	err := l.StreamChat(&types.ChatRequest{SessionID: "s-tool", Message: "read foo"}, w)
	require.NoError(t, err)

	body := w.Body.String()

	// Should emit tool_call SSE event
	assert.Contains(t, body, `"type":"tool_call"`)
	assert.Contains(t, body, `"id":"call_abc123"`)
	assert.Contains(t, body, `"name":"read_file"`)

	// Should emit tool_result SSE event
	assert.Contains(t, body, `"type":"tool_result"`)
	assert.Contains(t, body, `"id":"call_abc123"`)

	// Should emit final text token
	assert.Contains(t, body, "data: file processed\n\n")

	// Done event should be present
	assert.Contains(t, body, `"done":true`)

	// Verify AI was called twice (once for tool call, once after tool result)
	assert.Equal(t, 2, multiClient.callIdx)

	// Second call should include tool result message in history
	require.Len(t, multiClient.captures, 2)
	secondCallMsgs := multiClient.captures[1]
	// Find role=tool message
	foundToolMsg := false
	for _, msg := range secondCallMsgs {
		if msg.Role == openai.ChatMessageRoleTool {
			foundToolMsg = true
			break
		}
	}
	assert.True(t, foundToolMsg, "expected tool result message in history for second AI call")
}

// TestChatLogic_ApprovalGate: shell_run tool call that requires approval.
// A goroutine approves after a brief delay; stream should continue after approval.
func TestChatLogic_ApprovalGate(t *testing.T) {
	toolCallID := "call_shell_approve"
	argsJSON := `{"command":"ls -la"}`

	firstStream := makeToolCallsStream(toolCallID, "shell_run", argsJSON)
	secondStream := makeStopStream("done running")

	multiClient := &multiStreamAIClient{
		streams: []svc.StreamRecver{firstStream, secondStream},
	}

	// No allowlist means all commands require approval
	svcCtx := svc.NewServiceContextForTest(config.Config{
		Model: config.ModelConfig{
			SystemPrompt: "You are Jarvis.",
			Name:         "test-model",
		},
		TurnTimeoutSeconds: 60,
		MaxToolCalls:       5,
	}, multiClient, newMockConvStore())

	l := logic.NewChatLogic(context.Background(), svcCtx)

	// We need to intercept the approval_request SSE event to get the approvalId,
	// then resolve it. Use a pipe-based recorder.
	w := httptest.NewRecorder()

	// Spawn goroutine that waits for approval_request to appear in the buffer,
	// then resolves it. We poll the body briefly.
	done := make(chan error, 1)
	go func() {
		// Give logic time to emit approval_request and block
		time.Sleep(50 * time.Millisecond)
		// Scan body for approvalId
		body := w.Body.String()
		approvalID := extractApprovalID(body)
		if approvalID == "" {
			// Try a bit longer
			time.Sleep(100 * time.Millisecond)
			body = w.Body.String()
			approvalID = extractApprovalID(body)
		}
		if approvalID != "" {
			svcCtx.ApprovalStore.Resolve(approvalID, true)
		}
		done <- nil
	}()

	err := l.StreamChat(&types.ChatRequest{SessionID: "s-approve", Message: "run ls"}, w)
	<-done

	require.NoError(t, err)
	body := w.Body.String()

	// Should emit approval_request SSE event
	assert.Contains(t, body, `"type":"approval_request"`)
	assert.Contains(t, body, `"tool":"shell_run"`)
	assert.Contains(t, body, `"command":"ls -la"`)

	// Should emit tool_result (after approval)
	assert.Contains(t, body, `"type":"tool_result"`)

	// Final text should appear
	assert.Contains(t, body, "data: done running\n\n")
	assert.Contains(t, body, `"done":true`)
}

// TestChatLogic_ApprovalDenied: resolve with approved=false;
// verify ToolResult error is emitted and loop continues gracefully.
func TestChatLogic_ApprovalDenied(t *testing.T) {
	toolCallID := "call_shell_denied"
	argsJSON := `{"command":"rm -rf /"}`

	firstStream := makeToolCallsStream(toolCallID, "shell_run", argsJSON)
	secondStream := makeStopStream("noted, command denied")

	multiClient := &multiStreamAIClient{
		streams: []svc.StreamRecver{firstStream, secondStream},
	}

	svcCtx := svc.NewServiceContextForTest(config.Config{
		Model: config.ModelConfig{
			SystemPrompt: "You are Jarvis.",
			Name:         "test-model",
		},
		TurnTimeoutSeconds: 60,
		MaxToolCalls:       5,
	}, multiClient, newMockConvStore())

	l := logic.NewChatLogic(context.Background(), svcCtx)
	w := httptest.NewRecorder()

	done := make(chan error, 1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		body := w.Body.String()
		approvalID := extractApprovalID(body)
		if approvalID == "" {
			time.Sleep(100 * time.Millisecond)
			body = w.Body.String()
			approvalID = extractApprovalID(body)
		}
		if approvalID != "" {
			svcCtx.ApprovalStore.Resolve(approvalID, false)
		}
		done <- nil
	}()

	err := l.StreamChat(&types.ChatRequest{SessionID: "s-denied", Message: "delete everything"}, w)
	<-done

	require.NoError(t, err)
	body := w.Body.String()

	// Should emit approval_request event
	assert.Contains(t, body, `"type":"approval_request"`)

	// Should emit tool_result with error content (denial)
	assert.Contains(t, body, `"type":"tool_result"`)

	// Should contain error information in tool result
	assert.Contains(t, body, "denied")

	// Loop should continue after denial — done event should be present
	assert.Contains(t, body, `"done":true`)
}

// extractApprovalID parses the approvalId from SSE body containing approval_request events.
func extractApprovalID(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var event map[string]any
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event["type"] == "approval_request" {
			if id, ok := event["approvalId"].(string); ok {
				return id
			}
		}
	}
	return ""
}
