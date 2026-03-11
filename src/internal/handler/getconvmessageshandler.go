package handler

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"open-jarvis/internal/logic"
	"open-jarvis/internal/svc"
)

// GetConversationMessagesHandler returns an HTTP handler for GET /api/conversations/:id/messages.
// Always returns HTTP 200 with a JSON array (empty array when no messages found).
func GetConversationMessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `path:"id"`
		}
		if err := httpx.Parse(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		l := logic.NewGetConvMessagesLogic(r.Context(), svcCtx)
		msgs, err := l.GetConversationMessages(req.ID)
		if err != nil {
			logx.WithContext(r.Context()).Errorf("get conversation messages error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(msgs); err != nil {
			logx.WithContext(r.Context()).Errorf("encode messages: %v", err)
		}
	}
}
