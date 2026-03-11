package logic

import (
	"context"

	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// GetConvLogic handles fetching a single conversation.
type GetConvLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetConvLogic creates a new GetConvLogic instance.
func NewGetConvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConvLogic {
	return &GetConvLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetConversation returns the conversation with the given ID, or nil if not found.
// The caller should return a 404 when this returns nil.
func (l *GetConvLogic) GetConversation(id string) (*types.ConversationResponse, error) {
	conv, err := l.svcCtx.Store.GetConversation(id)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return nil, nil
	}
	return &types.ConversationResponse{
		ID:        conv.ID,
		Title:     conv.Title,
		CreatedAt: conv.CreatedAt,
		UpdatedAt: conv.UpdatedAt,
	}, nil
}
