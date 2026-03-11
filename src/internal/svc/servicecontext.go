package svc

import (
	"open-jarvis/internal/config"

	openai "github.com/sashabaranov/go-openai"
)

// ServiceContext holds all shared dependencies for the service.
type ServiceContext struct {
	Config    config.Config
	AIClient  *openai.Client
	ConvStore *ConvStore
}

// NewServiceContext creates a new ServiceContext with an OpenAI-compatible client
// configured from the provided config.
func NewServiceContext(c config.Config) *ServiceContext {
	cfg := openai.DefaultConfig(c.Model.APIKey)
	cfg.BaseURL = c.Model.BaseURL
	client := openai.NewClientWithConfig(cfg)

	return &ServiceContext{
		Config:    c,
		AIClient:  client,
		ConvStore: NewConvStore(),
	}
}
