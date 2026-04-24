package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
)

type TagRepository interface {
	ExistAll(ctx context.Context, ids []uuid.UUID) (bool, error)
	FindAllByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Tag, error)
	Create(ctx context.Context, tag *entity.Tag) (*entity.Tag, error)
	FindAll(ctx context.Context) ([]*entity.Tag, error)
}
