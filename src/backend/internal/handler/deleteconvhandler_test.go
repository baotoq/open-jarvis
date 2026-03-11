package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/rest/pathvar"

	"open-jarvis/internal/handler"
)

// injectPathParam returns a new request with the given path parameter injected
// into the context using go-zero's pathvar mechanism.
func injectPathParam(r *http.Request, key, value string) *http.Request {
	vars := map[string]string{key: value}
	return pathvar.WithVars(r, vars)
}

func TestDeleteConversation(t *testing.T) {
	store := newMockStore()
	_ = store.CreateConversation("conv-del", "To be deleted")

	svcCtx := newTestHandlerSvcCtx(store)
	h := handler.DeleteConversationHandler(svcCtx)

	r := httptest.NewRequest(http.MethodDelete, "/api/conversations/conv-del", nil)
	r = r.WithContext(context.Background())
	r = injectPathParam(r, "id", "conv-del")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())

	// Confirm deletion
	convs, _ := store.ListConversations()
	assert.Len(t, convs, 0)
}
