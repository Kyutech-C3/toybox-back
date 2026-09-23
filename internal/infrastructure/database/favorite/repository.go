package favorite

import (
	"context"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"
	"github.com/uptrace/bun"
)

type FavoriteRepository struct {
	db *bun.DB
}

func NewFavoriteRepository(db *bun.DB) *FavoriteRepository {
	return &FavoriteRepository{
		db: db,
	}
}

func (r *FavoriteRepository) Create(ctx context.Context, favorite *entity.Favorite) (*entity.Favorite, error) {
	dtoFavorite := dto.ToFavoriteDTO(favorite)

	_, err := r.db.NewInsert().Model(dtoFavorite).Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToCreateFavorite
	}
	return dtoFavorite.ToFavoriteEntity(), nil
}

func (r *FavoriteRepository) Delete(ctx context.Context, favorite *entity.Favorite) error {
	dtoFavorite := dto.ToFavoriteDTO(favorite)
	_, err := r.db.NewDelete().Model(dtoFavorite).Where("work_id = ? AND user_id = ?", dtoFavorite.WorkID, dtoFavorite.UserID).Exec(ctx)
	if err != nil {
		return domainerrors.ErrFailedToDeleteFavorite
	}
	return nil
}

func (r *FavoriteRepository) CountByWorkID(ctx context.Context, workID uuid.UUID) (int, error) {
	total, err := r.db.NewSelect().Model(&dto.Favorite{}).Where("work_id = ?", workID).Count(ctx)
	if err != nil {
		return 0, domainerrors.ErrFailedToCountFavoritesByWorkID
	}
	return total, nil
}

func (r *FavoriteRepository) Exists(ctx context.Context, favorite *entity.Favorite) (bool, error) {
	exists, err := r.db.NewSelect().Model(&dto.Favorite{}).Where("work_id = ? AND user_id = ?", favorite.WorkID, favorite.UserID).Exists(ctx)
	if err != nil {
		return false, domainerrors.ErrFailedToCheckFavoriteExists
	}
	return exists, nil
}

func (r *FavoriteRepository) FindFavoritedWorkIDs(ctx context.Context, userID uuid.UUID, workIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(workIDs) == 0 {
		return []uuid.UUID{}, nil
	}

	var favorites []dto.Favorite
	err := r.db.NewSelect().Model(&favorites).Column("work_id").Where("user_id = ? AND work_id IN (?)", userID, bun.In(workIDs)).Scan(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToFindFavoritedWorkIDs
	}

	workIDsResult := make([]uuid.UUID, 0, len(favorites))
	for _, favorite := range favorites {
		workIDsResult = append(workIDsResult, favorite.WorkID)
	}
	return workIDsResult, nil
}
