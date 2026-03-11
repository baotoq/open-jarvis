package config

import "github.com/zeromicro/go-zero/rest"

type ModelConfig struct {
	BaseURL      string `json:",default=http://localhost:11434/v1"`
	Name         string `json:",default=llama3.2"`
	APIKey       string `json:",optional"`
	SystemPrompt string `json:",default=You are Jarvis, a personal AI assistant. Be concise and helpful."`
}

type Config struct {
	rest.RestConf
	Model              ModelConfig
	MaxToolCalls       int `json:",default=10"`
	TurnTimeoutSeconds int `json:",default=60"`
}
