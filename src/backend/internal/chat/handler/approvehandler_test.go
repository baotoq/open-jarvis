package chat_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	handler "open-jarvis/internal/chat/handler"
	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// TestApproveHandler_Approved registers a channel, sends POST with approved=true,
// verifies channel receives true and response is 204 No Content.
func TestApproveHandler_Approved(t *testing.T) {
	approvalStore := svc.NewApprovalStore()
	svcCtx := &svc.ServiceContext{ApprovalStore: approvalStore}

	// Register a channel for the test approval ID
	approvalID := "test-approval-id-approved"
	ch := make(chan bool, 1)
	approvalStore.Register(approvalID, ch)

	body, err := json.Marshal(types.ApproveRequest{
		ApprovalID: approvalID,
		Approved:   true,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/chat/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ApproveHandler(svcCtx)(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Channel should have received true
	select {
	case approved := <-ch:
		assert.True(t, approved, "expected channel to receive true")
	default:
		t.Fatal("expected channel to have a value but it was empty")
	}
}

// TestApproveHandler_UnknownID sends POST with an unknown approval ID and verifies 404.
func TestApproveHandler_UnknownID(t *testing.T) {
	approvalStore := svc.NewApprovalStore()
	svcCtx := &svc.ServiceContext{ApprovalStore: approvalStore}

	body, err := json.Marshal(types.ApproveRequest{
		ApprovalID: "non-existent-id",
		Approved:   true,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/chat/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ApproveHandler(svcCtx)(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
