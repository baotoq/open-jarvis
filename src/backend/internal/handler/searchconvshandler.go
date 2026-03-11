package handler

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"open-jarvis/internal/logic"
	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// SearchConversationsHandler returns an HTTP handler for GET /api/conversations/search.
// Reads the search term from the ?q= query parameter.
// Returns 200 with a JSON array (never null — empty query returns []).
func SearchConversationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchConvsRequest
		req.Query = r.URL.Query().Get("q")

		l := logic.NewSearchConvsLogic(r.Context(), svcCtx)
		results, err := l.Search(req.Query)
		if err != nil {
			logx.WithContext(r.Context()).Errorf("search conversations: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if results == nil {
			results = []types.SearchResult{}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(results); err != nil {
			logx.WithContext(r.Context()).Errorf("encode search results: %v", err)
		}
	}
}
