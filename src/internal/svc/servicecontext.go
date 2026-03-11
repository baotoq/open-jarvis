package svc

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
	"open-jarvis/internal/config"
)

// StreamRecver abstracts a single streaming response.
// Mirrors logic.StreamRecver to avoid import cycles.
type StreamRecver interface {
	Recv() (openai.ChatCompletionStreamResponse, error)
	Close() error
}

// AIStreamer is the dependency interface for the AI client.
type AIStreamer interface {
	CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (StreamRecver, error)
}

// ServiceContext holds all shared dependencies for the service.
type ServiceContext struct {
	Config    config.Config
	AIClient  AIStreamer
	ConvStore *ConvStore
}

// NewServiceContext creates a new ServiceContext with a real OpenAI-compatible client.
func NewServiceContext(c config.Config) *ServiceContext {
	// Apply default system prompt if not set in config
	if c.Model.SystemPrompt == "" {
		c.Model.SystemPrompt = config.DefaultSystemPrompt
	}
	cfg := openai.DefaultConfig(c.Model.APIKey)
	cfg.BaseURL = c.Model.BaseURL
	client := openai.NewClientWithConfig(cfg)
	return &ServiceContext{
		Config:    c,
		AIClient:  &realClient{client: client},
		ConvStore: NewConvStore(),
	}
}

// NewServiceContextWithClient creates a ServiceContext with a provided AI client.
// Used in tests to inject mock clients.
func NewServiceContextWithClient(c config.Config, client AIStreamer, store *ConvStore) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		AIClient:  client,
		ConvStore: store,
	}
}

// realClient wraps *openai.Client to satisfy AIStreamer.
type realClient struct {
	client *openai.Client
}

func (r *realClient) CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (StreamRecver, error) {
	return r.client.CreateChatCompletionStream(ctx, req)
}
