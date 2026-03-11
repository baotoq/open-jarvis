package svc

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
	"open-jarvis/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigStoreGet(t *testing.T) {
	initial := config.ModelConfig{Name: "gpt-4"}
	cs := NewConfigStore(initial, "")
	got := cs.Get()
	assert.Equal(t, "gpt-4", got.Name)
}

func TestConfigStoreUpdate(t *testing.T) {
	initial := config.ModelConfig{Name: "gpt-4"}
	cs := NewConfigStore(initial, "")

	updated := config.ModelConfig{Name: "gpt-4o", BaseURL: "https://api.openai.com/v1"}
	err := cs.Update(updated)
	require.NoError(t, err)

	got := cs.Get()
	assert.Equal(t, "gpt-4o", got.Name)
	assert.Equal(t, "https://api.openai.com/v1", got.BaseURL)
}

func TestConfigStoreYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	initial := `Host: 0.0.0.0
Port: 8888
DBPath: data/conversations.db
Model:
  BaseURL: http://localhost:11434/v1
  Name: llama3.2
`
	err := os.WriteFile(cfgPath, []byte(initial), 0644)
	require.NoError(t, err)

	cs := NewConfigStore(config.ModelConfig{
		BaseURL: "http://localhost:11434/v1",
		Name:    "llama3.2",
	}, cfgPath)

	updated := config.ModelConfig{
		BaseURL: "https://api.openai.com/v1",
		Name:    "gpt-4o",
	}
	err = cs.Update(updated)
	require.NoError(t, err)

	// Read back the file and verify with yaml.Unmarshal
	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)

	var raw map[string]any
	err = yaml.Unmarshal(data, &raw)
	require.NoError(t, err)

	// Non-Model fields preserved
	assert.Equal(t, "0.0.0.0", raw["Host"])
	assert.Equal(t, 8888, raw["Port"])
	assert.Equal(t, "data/conversations.db", raw["DBPath"])

	// Model fields updated
	model, ok := raw["Model"].(map[string]any)
	require.True(t, ok, "Model field should be a map")
	assert.Equal(t, "gpt-4o", model["Name"])
	assert.Equal(t, "https://api.openai.com/v1", model["BaseURL"])
}

func TestConfigStoreYAML_MissingFile(t *testing.T) {
	cs := NewConfigStore(config.ModelConfig{}, "/nonexistent/path/config.yaml")
	err := cs.Update(config.ModelConfig{Name: "test"})
	assert.Error(t, err)
}

func TestConfigStoreUpdate_Concurrent(t *testing.T) {
	t.Parallel()
	cs := NewConfigStore(config.ModelConfig{Name: "initial"}, "")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			_ = cs.Update(config.ModelConfig{Name: "model"})
		}(i)
	}
	wg.Wait()
	// No panic and no data race
}
