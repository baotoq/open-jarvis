package handler

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"open-jarvis/internal/logic"
	"open-jarvis/internal/svc"
)

// GetConfigHandler returns an HTTP handler for GET /api/config.
// Returns 503 if ConfigStore is not initialised, otherwise 200 with JSON body.
func GetConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svcCtx.ConfigStore == nil {
			http.Error(w, "config store not available", http.StatusServiceUnavailable)
			return
		}
		l := logic.NewGetConfigLogic(r.Context(), svcCtx)
		resp := l.GetConfig()
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logx.WithContext(r.Context()).Errorf("encode config: %v", err)
		}
	}
}
