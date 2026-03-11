package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, "http://localhost:11434/v1", cfg.Model.BaseURL)
	assert.Equal(t, "llama3.2", cfg.Model.Name)
	assert.Equal(t, 10, cfg.MaxToolCalls)
	assert.Equal(t, 60, cfg.TurnTimeoutSeconds)
}

func TestConfig_Defaults(t *testing.T) {
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

	// WorkspaceRoot defaults to "." via go-zero struct tag
	assert.Equal(t, ".", cfg.WorkspaceRoot)
	// ShellAllowlist and ShellDenylist are optional; zero value is nil
	assert.Nil(t, cfg.ShellAllowlist)
	assert.Nil(t, cfg.ShellDenylist)
}
