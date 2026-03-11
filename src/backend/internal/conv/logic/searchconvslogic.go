package conv

import (
	"context"

	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// ConvSearcher is the interface satisfied by SQLiteConvStore for conversation search.
// Defined here (consumer package) following the "interfaces belong to consumers" rule.
type ConvSearcher interface {
	SearchConversations(query string) ([]svc.SearchResult, error)
}

// SearchConvsLogic handles full-text search over conversations.
type SearchConvsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSearchConvsLogic creates a new SearchConvsLogic instance.
func NewSearchConvsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchConvsLogic {
	return &SearchConvsLogic{ctx: ctx, svcCtx: svcCtx}
}

// Search returns conversations whose messages match the given query string.
// If the store does not support search, or if query is empty, returns nil, nil.
func (l *SearchConvsLogic) Search(query string) ([]types.SearchResult, error) {
	searcher, ok := l.svcCtx.Store.(ConvSearcher)
	if !ok {
		return nil, nil
	}
	results, err := searcher.SearchConversations(query)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return nil, nil
	}
	out := make([]types.SearchResult, len(results))
	for i, r := range results {
		out[i] = types.SearchResult{
			ID:        r.ConversationID,
			Title:     r.Title,
			UpdatedAt: r.UpdatedAt,
			Snippet:   r.Snippet,
		}
	}
	return out, nil
}
