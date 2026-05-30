package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
)

type WorkRepository interface {
	GetAll(ctx context.Context, limit, offset int, tagIDs []uuid.UUID) ([]*entity.Work, int, error)
	GetAllPublic(ctx context.Context, limit, offset int, tagIDs []uuid.UUID) ([]*entity.Work, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Work, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, public bool) ([]*entity.Work, error)
	ExistsById(ctx context.Context, id uuid.UUID) (bool, error)
	Create(ctx context.Context, work *entity.Work) (*entity.Work, error)
	Update(ctx context.Context, work *entity.Work) (*entity.Work, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}
