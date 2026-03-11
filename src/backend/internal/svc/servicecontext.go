package svc

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
	openai "github.com/sashabaranov/go-openai"
	"open-jarvis/internal/config"
	"open-jarvis/internal/toolexec"
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
	Config        config.Config
	AIClient      AIStreamer
	Store         ConversationStore
	Executor      toolexec.Executor
	ApprovalStore *ApprovalStore
	ShellTool     *toolexec.ShellTool
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

	var store ConversationStore
	if c.DBPath != "" {
		dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", c.DBPath)
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
			log.Fatalf("open sqlite: %v", err)
		}
		store, err = NewSQLiteConvStore(db)
		if err != nil {
			log.Fatalf("migrate sqlite: %v", err)
		}
	} else {
		store = NewConvStore()
	}

	fileTool := toolexec.NewFileTool(c.WorkspaceRoot)
	shellTool := toolexec.NewShellTool(c.ShellAllowlist, c.ShellDenylist)
	registry := toolexec.NewRegistry()
	registry.Register("read_file", fileTool.ReadFile)
	registry.Register("write_file", fileTool.WriteFile)
	registry.Register("shell_run", shellTool.Run)

	return &ServiceContext{
		Config:        c,
		AIClient:      &realClient{client: client},
		Store:         store,
		Executor:      registry,
		ApprovalStore: NewApprovalStore(),
		ShellTool:     shellTool,
	}
}

// NewServiceContextForTest creates a ServiceContext with the provided AI client and store.
// Used in tests to inject mock clients while still wiring real tool infrastructure.
func NewServiceContextForTest(c config.Config, client AIStreamer, store ConversationStore) *ServiceContext {
	fileTool := toolexec.NewFileTool(c.WorkspaceRoot)
	shellTool := toolexec.NewShellTool(c.ShellAllowlist, c.ShellDenylist)
	registry := toolexec.NewRegistry()
	registry.Register("read_file", fileTool.ReadFile)
	registry.Register("write_file", fileTool.WriteFile)
	registry.Register("shell_run", shellTool.Run)
	return &ServiceContext{
		Config:        c,
		AIClient:      client,
		Store:         store,
		Executor:      registry,
		ApprovalStore: NewApprovalStore(),
		ShellTool:     shellTool,
	}
}

// NewServiceContextWithClient creates a ServiceContext with a provided AI client.
// Used in tests to inject mock clients.
func NewServiceContextWithClient(c config.Config, client AIStreamer, store ConversationStore) *ServiceContext {
	return &ServiceContext{
		Config:   c,
		AIClient: client,
		Store:    store,
	}
}

// realClient wraps *openai.Client to satisfy AIStreamer.
type realClient struct {
	client *openai.Client
}

func (r *realClient) CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (StreamRecver, error) {
	return r.client.CreateChatCompletionStream(ctx, req)
}
