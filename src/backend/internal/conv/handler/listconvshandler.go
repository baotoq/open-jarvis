package conv

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	convlogic "open-jarvis/internal/conv/logic"
	"open-jarvis/internal/svc"
)

// ListConversationsHandler returns an HTTP handler for GET /api/conversations.
// It returns a JSON array of conversations ordered newest-first.
func ListConversationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := convlogic.NewListConvsLogic(r.Context(), svcCtx)
		convs, err := l.ListConversations()
		if err != nil {
			logx.WithContext(r.Context()).Errorf("list conversations error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(convs); err != nil {
			logx.WithContext(r.Context()).Errorf("encode conversations: %v", err)
		}
	}
}
