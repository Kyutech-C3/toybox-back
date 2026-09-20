package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/domain/repository"
)

type IFavoriteUsecase interface {
	CreateFavorite(ctx context.Context, workID uuid.UUID, userID uuid.UUID) error
	DeleteFavorite(ctx context.Context, workID uuid.UUID, userID uuid.UUID) error
	CountFavoritesByWorkID(ctx context.Context, workID uuid.UUID) (int, error)
	IsFavorite(ctx context.Context, workID uuid.UUID, userID uuid.UUID) bool
}

type favoriteUsecase struct {
	favoriteRepo repository.FavoriteRepository
	workRepo     repository.WorkRepository
}

func NewFavoriteUsecase(favoriteRepo repository.FavoriteRepository, workRepo repository.WorkRepository) IFavoriteUsecase {
	return &favoriteUsecase{
		favoriteRepo: favoriteRepo,
		workRepo:     workRepo,
	}
}

func (uc *favoriteUsecase) CreateFavorite(ctx context.Context, workID uuid.UUID, userID uuid.UUID) error {
	work, err := uc.workRepo.GetByID(ctx, workID)
	if err != nil {
		return fmt.Errorf("failed to get work by ID %s: %w", workID.String(), err)
	}
	if work.Visibility == "draft" && work.UserID != userID {
		return domainerrors.ErrWorkNotViewable
	}

	favorite := entity.NewFavorite(workID, userID)
	exists := uc.favoriteRepo.Exists(ctx, favorite)
	if exists {
		return domainerrors.ErrFavoriteAlreadyExists
	}

	_, err = uc.favoriteRepo.Create(ctx, favorite)
	if err != nil {
		return fmt.Errorf("failed to create favorite: %w", err)
	}
	return nil
}

func (uc *favoriteUsecase) DeleteFavorite(ctx context.Context, workID uuid.UUID, userID uuid.UUID) error {
	favorite := entity.NewFavorite(workID, userID)
	exists := uc.favoriteRepo.Exists(ctx, favorite)
	if !exists {
		return domainerrors.ErrFavoriteNotFound
	}
	return uc.favoriteRepo.Delete(ctx, favorite)
}

func (uc *favoriteUsecase) CountFavoritesByWorkID(ctx context.Context, workID uuid.UUID) (int, error) {
	total, err := uc.favoriteRepo.CountByWorkID(ctx, workID)
	if err != nil {
		return 0, fmt.Errorf("failed to count favorites by work ID %s: %w", workID.String(), err)
	}
	return total, nil
}

func (uc *favoriteUsecase) IsFavorite(ctx context.Context, workID uuid.UUID, userID uuid.UUID) bool {
	favorite := entity.NewFavorite(workID, userID)
	return uc.favoriteRepo.Exists(ctx, favorite)
}
