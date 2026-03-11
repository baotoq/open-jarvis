package conv

import (
	"context"

	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// ListConvsLogic handles listing all conversations.
type ListConvsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListConvsLogic creates a new ListConvsLogic instance.
func NewListConvsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListConvsLogic {
	return &ListConvsLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListConversations returns all conversations ordered newest-first.
func (l *ListConvsLogic) ListConversations() ([]types.ConversationResponse, error) {
	convs, err := l.svcCtx.Store.ListConversations()
	if err != nil {
		return nil, err
	}
	result := make([]types.ConversationResponse, 0, len(convs))
	for _, c := range convs {
		result = append(result, types.ConversationResponse{
			ID:        c.ID,
			Title:     c.Title,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		})
	}
	return result, nil
}
