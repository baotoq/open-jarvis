package handler_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"

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
	_ context.Context,
	_ openai.ChatCompletionRequest,
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
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
	assert.Equal(t, "no-cache", resp.Header.Get("Cache-Control"))
}
