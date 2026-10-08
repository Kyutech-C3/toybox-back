package webhook

import (
	"context"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
)

type WebhookNotifierRepository struct {
	webhookURL string
}

func NewWebhookNotifierRepository() *WebhookNotifierRepository {
	return &WebhookNotifierRepository{
		webhookURL: config.WEBHOOK_URL,
	}
}

func (r *WebhookNotifierRepository) WebhookNotify(ctx context.Context, work *entity.Work) error {
	return nil
}
