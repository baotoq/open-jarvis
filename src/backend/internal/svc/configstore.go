package svc

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"

	"open-jarvis/internal/config"
)

// ConfigStore holds the mutable model configuration and persists changes to disk.
// Thread-safe for concurrent reads; writes are serialised.
// If cfgPath is empty, Update still updates the in-memory value but skips the file write.
type ConfigStore struct {
	mu      sync.RWMutex
	cfg     config.ModelConfig
	cfgPath string
}

// NewConfigStore creates a ConfigStore with the given initial config and config file path.
func NewConfigStore(initial config.ModelConfig, cfgPath string) *ConfigStore {
	return &ConfigStore{cfg: initial, cfgPath: cfgPath}
}

// Get returns the current model configuration. Safe for concurrent use.
func (cs *ConfigStore) Get() config.ModelConfig {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.cfg
}

// Update writes the new config to disk (if cfgPath is set) and updates in-memory state.
// The in-memory state is only updated on a successful disk write.
func (cs *ConfigStore) Update(updated config.ModelConfig) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if err := cs.writeYAML(updated); err != nil {
		return err
	}
	cs.cfg = updated
	return nil
}

// writeYAML reads the existing YAML file, updates only the "Model" key, and writes back.
// Non-Model fields are preserved via a raw map[string]any round-trip.
// If cfgPath is empty, this is a no-op.
func (cs *ConfigStore) writeYAML(m config.ModelConfig) error {
	if cs.cfgPath == "" {
		return nil
	}
	data, err := os.ReadFile(cs.cfgPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	if raw == nil {
		raw = make(map[string]any)
	}
	raw["Model"] = map[string]any{
		"BaseURL":      m.BaseURL,
		"Name":         m.Name,
		"APIKey":       m.APIKey,
		"SystemPrompt": m.SystemPrompt,
	}
	out, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(cs.cfgPath, out, 0644)
}
