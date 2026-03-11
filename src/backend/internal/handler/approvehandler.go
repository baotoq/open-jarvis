// Package handler contains HTTP handlers for the open-jarvis API.
package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// ApproveHandler handles POST /api/chat/approve, resolving a pending approval gate.
func ApproveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ApproveRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}
		if !svcCtx.ApprovalStore.Resolve(req.ApprovalID, req.Approved) {
			http.Error(w, "unknown approval ID", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
