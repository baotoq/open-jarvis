package handler

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"open-jarvis/internal/logic"
	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// UpdateConfigHandler returns an HTTP handler for PUT /api/config.
// Accepts a JSON body with UpdateConfigRequest fields.
// Returns 204 on success, 400 on parse error, 500 on update error.
func UpdateConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		l := logic.NewUpdateConfigLogic(r.Context(), svcCtx)
		if err := l.UpdateConfig(&req); err != nil {
			logx.WithContext(r.Context()).Errorf("update config: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
