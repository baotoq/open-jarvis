package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"open-jarvis/internal/logic"
	"open-jarvis/internal/svc"
)

// DeleteConversationHandler returns an HTTP handler for DELETE /api/conversations/:id.
// Returns 204 No Content on success.
func DeleteConversationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `path:"id"`
		}
		if err := httpx.Parse(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		l := logic.NewDeleteConvLogic(r.Context(), svcCtx)
		if err := l.DeleteConversation(req.ID); err != nil {
			logx.WithContext(r.Context()).Errorf("delete conversation error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
