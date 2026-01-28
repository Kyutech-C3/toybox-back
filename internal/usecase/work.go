package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/domain/repository"
)

type IWorkUseCase interface {
	GetAll(ctx context.Context, limit, page *int, userID uuid.UUID, tagIDs []uuid.UUID) ([]*entity.Work, int, int, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Work, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, authenticatedUserID uuid.UUID) ([]*entity.Work, error)
	CreateWork(ctx context.Context, title, description, visibility string, thumbnailAssetID uuid.UUID, assetIDs []uuid.UUID, urls []string, userID uuid.UUID, tagIDs []uuid.UUID, collaboratorIDs []uuid.UUID) (*entity.Work, error)
	UpdateWork(ctx context.Context, workID uuid.UUID, userID uuid.UUID, title *string, description *string, visibility *string, thumbnailAssetID *uuid.UUID, assetIDs *[]uuid.UUID, urls *[]string, tagIDs *[]uuid.UUID, collaboratorIDs *[]uuid.UUID) (*entity.Work, error)
	DeleteWork(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type workUseCase struct {
	workRepo  repository.WorkRepository
	tagRepo   repository.TagRepository
	assetRepo repository.AssetRepository
	userRepo  repository.UserRepository
}

func NewWorkUseCase(workRepo repository.WorkRepository, tagRepo repository.TagRepository, assetRepo repository.AssetRepository, userRepo repository.UserRepository) IWorkUseCase {
	return &workUseCase{
		workRepo:  workRepo,
		tagRepo:   tagRepo,
		assetRepo: assetRepo,
		userRepo:  userRepo,
	}
}

func (uc *workUseCase) GetAll(ctx context.Context, limit, page *int, userID uuid.UUID, tagIDs []uuid.UUID) ([]*entity.Work, int, int, int, error) {
	actualLimit := 20
	actualPage := 1
	if limit != nil {
		actualLimit = *limit
	}
	if page != nil {
		actualPage = *page
	}
	fmt.Println("actualLimit", actualLimit)
	fmt.Println("actualPage", actualPage)
	fmt.Println("userID", userID)

	offset := (actualPage - 1) * actualLimit
	if userID == uuid.Nil {
		works, total, err := uc.workRepo.GetAllPublic(ctx, actualLimit, offset, tagIDs)
		if err != nil {
			return nil, 0, 0, 0, fmt.Errorf("failed to get all works by user ID %s: %w", userID.String(), err)
		}
		return works, total, actualLimit, actualPage, nil
	}

	works, total, err := uc.workRepo.GetAll(ctx, actualLimit, offset, tagIDs)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("failed to get all works: %w", err)
	}
	return works, total, actualLimit, actualPage, nil
}

func (uc *workUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Work, error) {
	work, err := uc.workRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get work by ID %s: %w", id.String(), err)
	}
	return work, nil
}

func (uc *workUseCase) GetByUserID(ctx context.Context, userID uuid.UUID, authenticatedUserID uuid.UUID) ([]*entity.Work, error) {
	var public bool
	if authenticatedUserID == uuid.Nil {
		public = true
	} else {
		public = false
	}

	works, err := uc.workRepo.GetByUserID(ctx, userID, public)
	if err != nil {
		return nil, fmt.Errorf("failed to get works by user ID %s: %w", userID.String(), err)
	}
	return works, nil
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
	if thumbnailAssetID != nil {
		work.ThumbnailAssetID = *thumbnailAssetID
	}
	if assetIDs != nil {
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
