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
	AuditStore    *AuditStore
	ConfigStore   *ConfigStore
}

// NewServiceContext creates a new ServiceContext with a real OpenAI-compatible client.
func NewServiceContext(c config.Config, configPath string) *ServiceContext {
	// Apply default system prompt if not set in config
	if c.Model.SystemPrompt == "" {
		c.Model.SystemPrompt = config.DefaultSystemPrompt
	}
	cfg := openai.DefaultConfig(c.Model.APIKey)
	cfg.BaseURL = c.Model.BaseURL
	client := openai.NewClientWithConfig(cfg)

	var store ConversationStore
	var auditStore *AuditStore
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
		auditStore, err = NewAuditStore(db)
		if err != nil {
			log.Fatalf("audit migrate: %v", err)
		}
	} else {
		store = NewConvStore()
	}

	if c.BraveSearchAPIKey == "" {
		log.Println("warning: BraveSearchAPIKey not set — web_search will return errors")
	}

	fileTool := toolexec.NewFileTool(c.WorkspaceRoot)
	shellTool := toolexec.NewShellTool(c.ShellAllowlist, c.ShellDenylist)
	webFetchTool := toolexec.NewWebFetchTool(c.WebFetchTimeoutSeconds)
	webSearchTool := toolexec.NewWebSearchTool(c.BraveSearchAPIKey)
	registry := toolexec.NewRegistry()
	registry.Register("read_file", fileTool.ReadFile)
	registry.Register("write_file", fileTool.WriteFile)
	registry.Register("shell_run", shellTool.Run)
	registry.Register("web_fetch", webFetchTool.Fetch)
	registry.Register("web_search", webSearchTool.Search)

	return &ServiceContext{
		Config:        c,
		AIClient:      &realClient{client: client},
		Store:         store,
		Executor:      registry,
		ApprovalStore: NewApprovalStore(),
		ShellTool:     shellTool,
		AuditStore:    auditStore,
		ConfigStore:   NewConfigStore(c.Model, configPath),
	}
}

// NewServiceContextForTest creates a ServiceContext with the provided AI client and store.
// Used in tests to inject mock clients while still wiring real tool infrastructure.
// Registers web tools (with empty API key — web_search returns "not configured" error in tests).
// Wires an in-memory AuditStore so audit log calls don't panic.
func NewServiceContextForTest(c config.Config, client AIStreamer, store ConversationStore) *ServiceContext {
	fileTool := toolexec.NewFileTool(c.WorkspaceRoot)
	shellTool := toolexec.NewShellTool(c.ShellAllowlist, c.ShellDenylist)
	webFetchTool := toolexec.NewWebFetchTool(c.WebFetchTimeoutSeconds)
	webSearchTool := toolexec.NewWebSearchTool(c.BraveSearchAPIKey)
	registry := toolexec.NewRegistry()
	registry.Register("read_file", fileTool.ReadFile)
	registry.Register("write_file", fileTool.WriteFile)
	registry.Register("shell_run", shellTool.Run)
	registry.Register("web_fetch", webFetchTool.Fetch)
	registry.Register("web_search", webSearchTool.Search)

	db, _ := sql.Open("sqlite", ":memory:")
	auditStore, _ := NewAuditStore(db)

	return &ServiceContext{
		Config:        c,
		AIClient:      client,
		Store:         store,
		Executor:      registry,
		ApprovalStore: NewApprovalStore(),
		ShellTool:     shellTool,
		AuditStore:    auditStore,
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

// RebuildAIClient replaces the AIClient with a new real OpenAI-compatible client
// using the given apiKey and baseURL. Called after a config update.
func (s *ServiceContext) RebuildAIClient(apiKey, baseURL string) {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = baseURL
	s.AIClient = &realClient{client: openai.NewClientWithConfig(cfg)}
}
