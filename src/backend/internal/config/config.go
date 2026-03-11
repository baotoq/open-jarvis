package config

import "github.com/zeromicro/go-zero/rest"

// DefaultSystemPrompt is the default system prompt used when none is configured.
const DefaultSystemPrompt = "You are Jarvis, a personal AI assistant. Be concise and helpful."

type ModelConfig struct {
	BaseURL      string `json:",default=http://localhost:11434/v1"`
	Name         string `json:",default=llama3.2"`
	APIKey       string `json:",optional"`
	SystemPrompt string `json:",optional"`
}

type Config struct {
	rest.RestConf
	Model              ModelConfig
	MaxToolCalls       int      `json:",default=10"`
	TurnTimeoutSeconds int      `json:",default=60"`
	DBPath             string   `json:",default=data/conversations.db"`
	ShellAllowlist     []string `json:",optional"`
	ShellDenylist      []string `json:",optional"`
	WorkspaceRoot             string   `json:",default=."`
	BraveSearchAPIKey         string   `json:",optional"`
	WebFetchTimeoutSeconds    int      `json:",default=30"`
}
