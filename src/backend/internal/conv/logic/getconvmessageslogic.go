package conv

import (
	"context"

	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// GetConvMessagesLogic handles fetching messages for a conversation.
type GetConvMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetConvMessagesLogic creates a new GetConvMessagesLogic instance.
func NewGetConvMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConvMessagesLogic {
	return &GetConvMessagesLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetConversationMessages returns the messages for a given conversation ID.
// Returns an empty slice (never nil) when the conversation has no messages.
func (l *GetConvMessagesLogic) GetConversationMessages(id string) ([]types.MessageResponse, error) {
	msgs := l.svcCtx.Store.Get(id)
	result := make([]types.MessageResponse, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, types.MessageResponse{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}
	return result, nil
}
