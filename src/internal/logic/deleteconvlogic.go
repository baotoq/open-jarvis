package logic

import (
	"context"

	"open-jarvis/internal/svc"
)

// DeleteConvLogic handles deleting a conversation.
type DeleteConvLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteConvLogic creates a new DeleteConvLogic instance.
func NewDeleteConvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteConvLogic {
	return &DeleteConvLogic{ctx: ctx, svcCtx: svcCtx}
}

// DeleteConversation removes a conversation by ID.
func (l *DeleteConvLogic) DeleteConversation(id string) error {
	return l.svcCtx.Store.DeleteConversation(id)
}
