package repository

import (
	"context"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
)

type WorkNotifier interface {
	NotifyCreated(ctx context.Context, work *entity.Work) error
}
