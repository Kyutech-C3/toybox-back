package usecase

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/domain/repository"
)

type IWorkUseCase interface {
	GetAll(ctx context.Context, limit, page *int, userID uuid.UUID, tagIDs []uuid.UUID) ([]*entity.Work, int, int, int, map[uuid.UUID]bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Work, error)
	GetByUserID(ctx context.Context, limit, page *int, userID uuid.UUID, authenticatedUserID uuid.UUID) ([]*entity.Work, int, int, int, map[uuid.UUID]bool, error)
	CreateWork(ctx context.Context, title, description, visibility string, thumbnailAssetID uuid.UUID, assetIDs []uuid.UUID, urls []string, userID uuid.UUID, tagIDs []uuid.UUID, collaboratorIDs []uuid.UUID) (*entity.Work, error)
	UpdateWork(ctx context.Context, workID uuid.UUID, userID uuid.UUID, title *string, description *string, visibility *string, thumbnailAssetID *uuid.UUID, assetIDs *[]uuid.UUID, urls *[]string, tagIDs *[]uuid.UUID, collaboratorIDs *[]uuid.UUID) (*entity.Work, error)
	DeleteWork(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type workUseCase struct {
	workRepo     repository.WorkRepository
	tagRepo      repository.TagRepository
	assetRepo    repository.AssetRepository
	userRepo     repository.UserRepository
	favoriteRepo repository.FavoriteRepository
}

func NewWorkUseCase(workRepo repository.WorkRepository, tagRepo repository.TagRepository, assetRepo repository.AssetRepository, userRepo repository.UserRepository, favoriteRepo repository.FavoriteRepository) IWorkUseCase {
	return &workUseCase{
		workRepo:     workRepo,
		tagRepo:      tagRepo,
		assetRepo:    assetRepo,
		userRepo:     userRepo,
		favoriteRepo: favoriteRepo,
	}
}

func (uc *workUseCase) favoritedWorkIDs(ctx context.Context, viewerID uuid.UUID, works []*entity.Work) (map[uuid.UUID]bool, error) {
	if viewerID == uuid.Nil || len(works) == 0 {
		return map[uuid.UUID]bool{}, nil
	}

	workIDs := make([]uuid.UUID, len(works))
	for i, work := range works {
		workIDs[i] = work.ID
	}

	favoritedIDs, err := uc.favoriteRepo.FindFavoritedWorkIDs(ctx, viewerID, workIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find favorited work ids: %w", err)
	}

	favoritedSet := make(map[uuid.UUID]bool, len(favoritedIDs))
	for _, workID := range favoritedIDs {
		favoritedSet[workID] = true
	}
	return favoritedSet, nil
}

func (uc *workUseCase) GetAll(ctx context.Context, limit, page *int, userID uuid.UUID, tagIDs []uuid.UUID) ([]*entity.Work, int, int, int, map[uuid.UUID]bool, error) {
	actualLimit := 20
	actualPage := 1
	if limit != nil {
		actualLimit = *limit
	}
	if page != nil {
		actualPage = *page
	}
	offset := (actualPage - 1) * actualLimit
	if userID == uuid.Nil {
		works, total, err := uc.workRepo.GetAllPublic(ctx, actualLimit, offset, tagIDs)
		if err != nil {
			return nil, 0, 0, 0, nil, fmt.Errorf("failed to get all works by user ID %s: %w", userID.String(), err)
		}
		favoritedWorkIDs, err := uc.favoritedWorkIDs(ctx, userID, works)
		if err != nil {
			return nil, 0, 0, 0, nil, err
		}
		return works, total, actualLimit, actualPage, favoritedWorkIDs, nil
	}

	works, total, err := uc.workRepo.GetAll(ctx, actualLimit, offset, tagIDs)
	if err != nil {
		return nil, 0, 0, 0, nil, fmt.Errorf("failed to get all works: %w", err)
	}
	favoritedWorkIDs, err := uc.favoritedWorkIDs(ctx, userID, works)
	if err != nil {
		return nil, 0, 0, 0, nil, err
	}
	return works, total, actualLimit, actualPage, favoritedWorkIDs, nil
}

func (uc *workUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Work, error) {
	work, err := uc.workRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get work by ID %s: %w", id.String(), err)
	}
	return work, nil
}

func (uc *workUseCase) GetByUserID(ctx context.Context, limit, page *int, userID uuid.UUID, authenticatedUserID uuid.UUID) ([]*entity.Work, int, int, int, map[uuid.UUID]bool, error) {
	actualLimit := 20
	actualPage := 1
	if limit != nil {
		actualLimit = *limit
	}
	if page != nil {
		actualPage = *page
	}
	offset := (actualPage - 1) * actualLimit

	includePrivate := false
	includeDraft := false
	if authenticatedUserID != uuid.Nil {
		includePrivate = true
		if authenticatedUserID == userID {
			includeDraft = true
		}
	}

	works, total, err := uc.workRepo.GetByUserID(ctx, userID, includePrivate, includeDraft, actualLimit, offset)
	if err != nil {
		return nil, 0, 0, 0, nil, fmt.Errorf("failed to get works by user ID %s: %w", userID.String(), err)
	}
	favoritedWorkIDs, err := uc.favoritedWorkIDs(ctx, authenticatedUserID, works)
	if err != nil {
		return nil, 0, 0, 0, nil, err
	}
	return works, total, actualLimit, actualPage, favoritedWorkIDs, nil
}

func (uc *workUseCase) CreateWork(ctx context.Context, title, description, visibility string, thumbnailAssetID uuid.UUID, assetIDs []uuid.UUID, urls []string, userID uuid.UUID, tagIDs []uuid.UUID, collaboratorIDs []uuid.UUID) (*entity.Work, error) {
	if title == "" {
		return nil, domainerrors.ErrInvalidTitle
	}
	if description == "" {
		return nil, domainerrors.ErrInvalidDescription
	}
	if visibility == "" {
		return nil, domainerrors.ErrInvalidVisibility
	}
	if len(tagIDs) == 0 {
		return nil, domainerrors.ErrInvalidTagIDs
	}

	var tags []*entity.Tag
	var err error

	exists, err := uc.tagRepo.ExistAll(ctx, tagIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to check tag existence: %w", err)
	}
	if !exists {
		return nil, domainerrors.ErrTagNotFound
	}

	tags, err = uc.tagRepo.FindAllByIDs(ctx, tagIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find tags by ids: %w", err)
	}

	var collaborators []*entity.User
	if len(collaboratorIDs) > 0 {
		for _, collaboratorID := range collaboratorIDs {
			// オーナー自身を共同制作者として追加できないようにする
			if collaboratorID == userID {
				return nil, domainerrors.ErrOwnerCannotBeCollaborator
			}
			user, err := uc.userRepo.GetByID(ctx, collaboratorID)
			if err != nil {
				return nil, fmt.Errorf("failed to get collaborator by ID %s: %w", collaboratorID.String(), err)
			}
			collaborators = append(collaborators, user)
		}
	}

	ownedAssetIDs := append([]uuid.UUID{thumbnailAssetID}, assetIDs...)
	ownsAllAssets, err := uc.assetRepo.ExistAllByUserID(ctx, ownedAssetIDs, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check asset ownership: %w", err)
	}
	if !ownsAllAssets {
		return nil, domainerrors.ErrAssetNotFound
	}

	assets := make([]*entity.Asset, len(assetIDs))
	for i, assetID := range assetIDs {
		assets[i] = &entity.Asset{
			ID: assetID,
		}
	}

	urlPointers := make([]*string, len(urls))
	for i, url := range urls {
		urlPointers[i] = &url
	}

	work := entity.NewWork(title, description, userID, visibility, thumbnailAssetID, assets, urlPointers, tagIDs, tags)
	work.Collaborators = collaborators

	createdWork, err := uc.workRepo.Create(ctx, work)
	if err != nil {
		return nil, fmt.Errorf("failed to create work: %w", err)
	}
	return createdWork, nil
}

func (uc *workUseCase) UpdateWork(ctx context.Context, workID uuid.UUID, userID uuid.UUID, title *string, description *string, visibility *string, thumbnailAssetID *uuid.UUID, assetIDs *[]uuid.UUID, urls *[]string, tagIDs *[]uuid.UUID, collaboratorIDs *[]uuid.UUID) (*entity.Work, error) {
	work, err := uc.workRepo.GetByID(ctx, workID)
	if err != nil {
		return nil, fmt.Errorf("failed to get work by ID %s: %w", workID.String(), err)
	}

	if work.UserID != userID {
		return nil, domainerrors.ErrWorkNotOwnedByUser
	}

	if title != nil {
		work.Title = *title
	}
	if description != nil {
		work.Description = *description
	}
	if visibility != nil {
		work.Visibility = *visibility
	}
	if thumbnailAssetID != nil && *thumbnailAssetID == uuid.Nil {
		return nil, domainerrors.ErrInvalidThumbnailAssetID
	}

	verifyAssetIDs := make([]uuid.UUID, 0, 1)
	if thumbnailAssetID != nil {
		verifyAssetIDs = append(verifyAssetIDs, *thumbnailAssetID)
	}
	if assetIDs != nil {
		verifyAssetIDs = append(verifyAssetIDs, *assetIDs...)
	}
	if len(verifyAssetIDs) > 0 {
		ownsAllAssets, err := uc.assetRepo.ExistAllByUserID(ctx, verifyAssetIDs, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check asset ownership: %w", err)
		}
		if !ownsAllAssets {
			return nil, domainerrors.ErrAssetNotFound
		}
	}

	if thumbnailAssetID != nil {
		work.ThumbnailAssetID = *thumbnailAssetID
	}
	var removedAssets []*entity.Asset
	if assetIDs != nil {
		for _, oldAsset := range work.Assets {
			if oldAsset.ID != work.ThumbnailAssetID && !slices.Contains(*assetIDs, oldAsset.ID) {
				removedAssets = append(removedAssets, oldAsset)
			}
		}

		assets := make([]*entity.Asset, len(*assetIDs))
		for i, assetID := range *assetIDs {
			assets[i] = &entity.Asset{
				ID: assetID,
			}
		}
		work.Assets = assets
	}
	if urls != nil {
		urlPointers := make([]*string, len(*urls))
		for i, url := range *urls {
			urlPointers[i] = &url
		}
		work.URLs = urlPointers
	}
	if tagIDs != nil {
		exists, err := uc.tagRepo.ExistAll(ctx, *tagIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to check tag existence: %w", err)
		}
		if !exists {
			return nil, domainerrors.ErrTagNotFound
		}
		tags, err := uc.tagRepo.FindAllByIDs(ctx, *tagIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to find tags by ids: %w", err)
		}
		work.Tags = tags
		work.TagIDs = *tagIDs
	}
	if collaboratorIDs != nil {
		var collaborators []*entity.User
		if len(*collaboratorIDs) > 0 {
			for _, collaboratorID := range *collaboratorIDs {
				// オーナー自身を共同制作者として追加できないようにする
				if collaboratorID == userID {
					return nil, domainerrors.ErrOwnerCannotBeCollaborator
				}
				user, err := uc.userRepo.GetByID(ctx, collaboratorID)
				if err != nil {
					return nil, fmt.Errorf("failed to get collaborator by ID %s: %w", collaboratorID.String(), err)
				}
				collaborators = append(collaborators, user)
			}
		}
		work.Collaborators = collaborators
	}

	updatedWork, err := uc.workRepo.Update(ctx, work)
	if err != nil {
		return nil, fmt.Errorf("failed to update work: %w", err)
	}

	for _, asset := range removedAssets {
		if err := uc.assetRepo.DeleteFile(ctx, asset.URL); err != nil {
			return nil, fmt.Errorf("failed to delete removed asset file: %w", err)
		}
	}

	return updatedWork, nil
}

func (uc *workUseCase) DeleteWork(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	work, err := uc.workRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get work by ID %s: %w", id.String(), err)
	}

	if work.UserID != userID {
		return domainerrors.ErrWorkNotOwnedByUser
	}

	err = uc.workRepo.Delete(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete work %s: %w", id.String(), err)
	}

	for _, asset := range work.Assets {
		if err := uc.assetRepo.DeleteFile(ctx, asset.URL); err != nil {
			return fmt.Errorf("failed to delete asset file: %w", err)
		}
	}

	return nil
}
