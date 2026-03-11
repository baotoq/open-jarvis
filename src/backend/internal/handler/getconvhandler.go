package handler

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"open-jarvis/internal/logic"
	"open-jarvis/internal/svc"
)

// GetConversationHandler returns an HTTP handler for GET /api/conversations/:id.
// Returns 404 when the conversation does not exist.
func GetConversationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `path:"id"`
		}
		if err := httpx.Parse(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		l := logic.NewGetConvLogic(r.Context(), svcCtx)
		conv, err := l.GetConversation(req.ID)
		if err != nil {
			logx.WithContext(r.Context()).Errorf("get conversation error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if conv == nil {
			http.Error(w, "conversation not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(conv); err != nil {
			logx.WithContext(r.Context()).Errorf("encode conversation: %v", err)
		}
	}
}
