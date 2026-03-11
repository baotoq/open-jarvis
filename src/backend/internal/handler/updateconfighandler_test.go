package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"open-jarvis/internal/config"
	"open-jarvis/internal/handler"
)

func TestUpdateConfigHandler(t *testing.T) {
	mc := config.ModelConfig{
		BaseURL: "http://localhost:11434/v1",
		Name:    "llama3.2",
	}
	svcCtx := newTestSvcCtxWithConfigStore(mc)
	h := handler.UpdateConfigHandler(svcCtx)

	body := map[string]string{
		"baseURL":      "http://newhost:11434/v1",
		"name":         "mistral",
		"apiKey":       "new-key",
		"systemPrompt": "New prompt.",
	}
	b, _ := json.Marshal(body)
	r := httptest.NewRequest(http.MethodPut, "/api/config", bytes.NewReader(b))
	r = r.WithContext(context.Background())
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify config was persisted in-memory
	updated := svcCtx.ConfigStore.Get()
	assert.Equal(t, "http://newhost:11434/v1", updated.BaseURL)
	assert.Equal(t, "mistral", updated.Name)
	assert.Equal(t, "new-key", updated.APIKey)
	assert.Equal(t, "New prompt.", updated.SystemPrompt)
}

func TestUpdateConfigHandler_BadBody(t *testing.T) {
	mc := config.ModelConfig{}
	svcCtx := newTestSvcCtxWithConfigStore(mc)
	h := handler.UpdateConfigHandler(svcCtx)

	r := httptest.NewRequest(http.MethodPut, "/api/config", strings.NewReader("not json"))
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
