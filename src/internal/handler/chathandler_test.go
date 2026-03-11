package handler_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"open-jarvis/internal/config"
	"open-jarvis/internal/handler"
	"open-jarvis/internal/svc"
)

// simpleStream provides a simple mock stream for handler tests.
type simpleStream struct {
	tokens []string
	pos    int
}

func (s *simpleStream) Recv() (openai.ChatCompletionStreamResponse, error) {
	if s.pos >= len(s.tokens) {
		return openai.ChatCompletionStreamResponse{}, io.EOF
	}
	token := s.tokens[s.pos]
	s.pos++
	return openai.ChatCompletionStreamResponse{
		Choices: []openai.ChatCompletionStreamChoice{
			{Delta: openai.ChatCompletionStreamChoiceDelta{Content: token}},
		},
	}, nil
}
func (s *simpleStream) Close() error { return nil }

type simpleClient struct{}

func (c *simpleClient) CreateChatCompletionStream(
	ctx context.Context,
	req openai.ChatCompletionRequest,
) (svc.StreamRecver, error) {
	return &simpleStream{tokens: []string{"hello"}}, nil
}

func TestChatStreamHandlerHeaders(t *testing.T) {
	client := &simpleClient{}
	svcCtx := svc.NewServiceContextWithClient(config.Config{
		Model: config.ModelConfig{
			SystemPrompt: "You are Jarvis.",
			Name:         "test-model",
		},
		TurnTimeoutSeconds: 60,
	}, client, svc.NewConvStore())

	h := handler.ChatStreamHandler(svcCtx)
	body := `{"sessionId":"s1","message":"hi"}`
	req := httptest.NewRequest(http.MethodPost, "/api/chat/stream", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h(w, req)

	resp := w.Result()
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected Content-Type: text/event-stream, got %q", ct)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("expected Cache-Control: no-cache, got %q", cc)
	}
}
