package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
)

type FavoriteRepository interface {
	Create(ctx context.Context, favorite *entity.Favorite) (*entity.Favorite, error)
	Delete(ctx context.Context, favorite *entity.Favorite) error
	CountByWorkID(ctx context.Context, workID uuid.UUID) (int, error)
	Exists(ctx context.Context, favorite *entity.Favorite) bool
	FindFavoritedWorkIDs(ctx context.Context, userID uuid.UUID, workIDs []uuid.UUID) ([]uuid.UUID, error)
}
