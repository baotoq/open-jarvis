package logic

import (
	"context"

	"open-jarvis/internal/config"
	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// UpdateConfigLogic handles updating the model configuration.
type UpdateConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateConfigLogic creates a new UpdateConfigLogic instance.
func NewUpdateConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateConfigLogic {
	return &UpdateConfigLogic{ctx: ctx, svcCtx: svcCtx}
}

// UpdateConfig persists the new model configuration and rebuilds the AIClient.
func (l *UpdateConfigLogic) UpdateConfig(req *types.UpdateConfigRequest) error {
	updated := config.ModelConfig{
		BaseURL:      req.BaseURL,
		Name:         req.Name,
		APIKey:       req.APIKey,
		SystemPrompt: req.SystemPrompt,
	}
	if err := l.svcCtx.ConfigStore.Update(updated); err != nil {
		return err
	}
	l.svcCtx.RebuildAIClient(req.APIKey, req.BaseURL)
	return nil
}
