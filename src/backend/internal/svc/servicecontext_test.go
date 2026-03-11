package svc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"open-jarvis/internal/config"
	"open-jarvis/internal/svc"
)

// TestNewServiceContextForTest verifies that NewServiceContextForTest wires
// AuditStore and web tools correctly.
func TestNewServiceContextForTest(t *testing.T) {
	c := config.Config{}
	store := svc.NewConvStore()
	// Use a minimal mock client — we just need ServiceContext fields populated.
	svcCtx := svc.NewServiceContextForTest(c, nil, store)

	t.Run("AuditStore is non-nil", func(t *testing.T) {
		require.NotNil(t, svcCtx.AuditStore)
	})

	t.Run("Executor has web_fetch tool", func(t *testing.T) {
		// Execute web_fetch with empty URL — should return ToolResult with error
		// not "unknown tool: web_fetch"
		result := svcCtx.Executor.Execute(t.Context(), "web_fetch", `{"url":""}`)
		assert.NotEqual(t, "unknown tool: web_fetch", result.Error,
			"web_fetch must be registered in NewServiceContextForTest")
	})

	t.Run("Executor has web_search tool", func(t *testing.T) {
		// Execute web_search with empty key — should return a "not configured" error
		// not "unknown tool: web_search"
		result := svcCtx.Executor.Execute(t.Context(), "web_search", `{"query":"test"}`)
		assert.NotEqual(t, "unknown tool: web_search", result.Error,
			"web_search must be registered in NewServiceContextForTest")
	})
}

// TestNewServiceContextWithClient_AuditStoreNil verifies that the legacy
// NewServiceContextWithClient leaves AuditStore nil (nil-guard path).
func TestNewServiceContextWithClient_AuditStoreNil(t *testing.T) {
	c := config.Config{}
	store := svc.NewConvStore()
	svcCtx := svc.NewServiceContextWithClient(c, nil, store)
	assert.Nil(t, svcCtx.AuditStore,
		"NewServiceContextWithClient must leave AuditStore nil")
}
