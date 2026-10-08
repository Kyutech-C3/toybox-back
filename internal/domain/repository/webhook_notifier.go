package repository

import (
	"context"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
)

type WebhookNotifierRepository interface {
	WebhookNotify(ctx context.Context, work *entity.Work) error
}
