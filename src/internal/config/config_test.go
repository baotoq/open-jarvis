package config_test

import (
	"os"
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
	"open-jarvis/internal/config"
)

func TestConfigDefaults(t *testing.T) {
	// Write a minimal YAML with only required RestConf fields
	yaml := `Name: test
Host: 0.0.0.0
Port: 8888
`
	f, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(yaml); err != nil {
		t.Fatal(err)
	}
	f.Close()

	var cfg config.Config
	conf.MustLoad(f.Name(), &cfg)

	if cfg.Model.BaseURL != "http://localhost:11434/v1" {
		t.Errorf("expected BaseURL=http://localhost:11434/v1, got %q", cfg.Model.BaseURL)
	}
	if cfg.Model.Name != "llama3.2" {
		t.Errorf("expected Name=llama3.2, got %q", cfg.Model.Name)
	}
	if cfg.MaxToolCalls != 10 {
		t.Errorf("expected MaxToolCalls=10, got %d", cfg.MaxToolCalls)
	}
	if cfg.TurnTimeoutSeconds != 60 {
		t.Errorf("expected TurnTimeoutSeconds=60, got %d", cfg.TurnTimeoutSeconds)
	}
}
