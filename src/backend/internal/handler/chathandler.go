package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/core/logx"
	"open-jarvis/internal/logic"
	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// ChatStreamHandler returns an HTTP handler for the POST /api/chat/stream endpoint.
// It sets SSE headers, parses the request, and delegates to ChatLogic.StreamChat.
func ChatStreamHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		var req types.ChatRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewChatLogic(r.Context(), svcCtx)
		if err := l.StreamChat(&req, w); err != nil {
			// Stream may be partially written — log but do not re-write headers
			logx.WithContext(r.Context()).Errorf("stream chat error: %v", err)
		}
	}
}
