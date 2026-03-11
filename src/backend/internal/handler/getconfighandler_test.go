package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"open-jarvis/internal/config"
	"open-jarvis/internal/handler"
	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

func newTestSvcCtxWithConfigStore(mc config.ModelConfig) *svc.ServiceContext {
	svcCtx := svc.NewServiceContextWithClient(config.Config{}, nil, newMockStore())
	svcCtx.ConfigStore = svc.NewConfigStore(mc, "")
	return svcCtx
}

func TestGetConfigHandler(t *testing.T) {
	mc := config.ModelConfig{
		BaseURL:      "http://localhost:11434/v1",
		Name:         "llama3.2",
		APIKey:       "test-key",
		SystemPrompt: "You are helpful.",
	}
	svcCtx := newTestSvcCtxWithConfigStore(mc)
	h := handler.GetConfigHandler(svcCtx)

	r := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp types.ConfigResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, mc.BaseURL, resp.BaseURL)
	assert.Equal(t, mc.Name, resp.Name)
	assert.Equal(t, mc.APIKey, resp.APIKey)
	assert.Equal(t, mc.SystemPrompt, resp.SystemPrompt)
}

func TestGetConfigHandler_NilStore(t *testing.T) {
	svcCtx := svc.NewServiceContextWithClient(config.Config{}, nil, newMockStore())
	// ConfigStore is nil by default in NewServiceContextWithClient
	h := handler.GetConfigHandler(svcCtx)

	r := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
