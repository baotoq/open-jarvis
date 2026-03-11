package logic

import (
	"context"

	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// GetConfigLogic handles retrieving the current model configuration.
type GetConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetConfigLogic creates a new GetConfigLogic instance.
func NewGetConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConfigLogic {
	return &GetConfigLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetConfig returns the current model configuration from the ConfigStore.
func (l *GetConfigLogic) GetConfig() *types.ConfigResponse {
	m := l.svcCtx.ConfigStore.Get()
	return &types.ConfigResponse{
		BaseURL:      m.BaseURL,
		Name:         m.Name,
		APIKey:       m.APIKey,
		SystemPrompt: m.SystemPrompt,
	}
}
