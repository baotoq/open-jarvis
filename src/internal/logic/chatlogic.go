package logic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// StreamRecver abstracts the go-openai streaming response for testability.
// This mirrors svc.StreamRecver to allow mock injection in tests without import cycles.
type StreamRecver interface {
	Recv() (openai.ChatCompletionStreamResponse, error)
	Close() error
}

// ChatLogic handles the streaming LLM call and conversation management.
type ChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewChatLogic creates a new ChatLogic instance.
func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{ctx: ctx, svcCtx: svcCtx}
}

// StreamChat performs the streaming LLM call, writing SSE tokens to w.
// It prepends the system prompt for new sessions, appends conversation history,
// and stores the assembled response in ConvStore after completion.
func (l *ChatLogic) StreamChat(req *types.ChatRequest, w http.ResponseWriter) error {
	ctx, cancel := context.WithTimeout(l.ctx,
		time.Duration(l.svcCtx.Config.TurnTimeoutSeconds)*time.Second)
	defer cancel()

	// Build message history with system prompt for new sessions
	history := l.svcCtx.Store.Get(req.SessionId)
	if len(history) == 0 {
		history = append(history, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: l.svcCtx.Config.Model.SystemPrompt,
		})
	}
	history = append(history, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Message,
	})

	// Assert flusher before making the LLM call
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported: ResponseWriter does not implement http.Flusher")
	}

	stream, err := l.svcCtx.AIClient.CreateChatCompletionStream(ctx,
		openai.ChatCompletionRequest{
			Model:    l.svcCtx.Config.Model.Name,
			Messages: history,
		})
	if err != nil {
		fmt.Fprintf(w, "data: [ERROR] %s\n\n", err.Error())
		flusher.Flush()
		return fmt.Errorf("create stream: %w", err)
	}
	defer stream.Close()

	var fullResponse strings.Builder

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Fprintf(w, "data: [ERROR] %s\n\n", err.Error())
			flusher.Flush()
			return fmt.Errorf("stream recv: %w", err)
		}
		if len(resp.Choices) == 0 {
			continue
		}
		token := resp.Choices[0].Delta.Content
		if token == "" {
			continue
		}
		fullResponse.WriteString(token)
		fmt.Fprintf(w, "data: %s\n\n", token)
		flusher.Flush()
	}

	// Persist the full conversation turn to the store
	history = append(history, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: fullResponse.String(),
	})
	l.svcCtx.Store.Set(req.SessionId, history)
	return nil
}
