package usecase

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/domain/repository"
)

type IUserUseCase interface {
	GetAllUser(ctx context.Context) ([]*entity.User, error)
	GetByUserID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, displayName *string, profile *string, xUsername *string, githubID *string, avatar *multipart.FileHeader) (*entity.User, error)
}

type userUseCase struct {
	repo      repository.UserRepository
	assetRepo repository.AssetRepository
}

var allowedAvatarExtensions = map[string]struct{}{
	"png":  {},
	"jpg":  {},
	"jpeg": {},
	"gif":  {},
	"webp": {},
}

func NewUserUseCase(repo repository.UserRepository, assetRepo repository.AssetRepository) IUserUseCase {
	return &userUseCase{
		repo:      repo,
		assetRepo: assetRepo,
	}
}

func (u *userUseCase) GetAllUser(ctx context.Context) ([]*entity.User, error) {
	return u.repo.GetAll(ctx)
}

func (u *userUseCase) GetByUserID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID %s: %w", id.String(), err)
	}
	return user, nil
}

func (u *userUseCase) UpdateUser(ctx context.Context, userID uuid.UUID, displayName *string, profile *string, xUsername *string, githubID *string, avatar *multipart.FileHeader) (*entity.User, error) {
	var avatarExtension string
	if avatar != nil {
		extension, err := extractExtension(avatar.Filename)
		if err != nil {
			return nil, err
		}
		if _, ok := allowedAvatarExtensions[extension]; !ok {
			return nil, domainerrors.ErrUnsupportedFileType
		}
		avatarExtension = extension
	}

	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID %s: %w", userID.String(), err)
	}
	if displayName != nil {
		user.DisplayName = *displayName
	}
	if profile != nil {
		user.Profile = *profile
	}
	if xUsername != nil {
		user.XUsername = *xUsername
	}
	if githubID != nil {
		user.GithubID = *githubID
	}
	oldAvatarURL := user.AvatarURL
	var newAvatarURL *string
	if avatar != nil {
		newAvatarURL, err = u.assetRepo.UploadAvatarFile(ctx, avatar, avatarExtension)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", domainerrors.ErrFailedToUploadFile, err)
		}
		user.AvatarURL = *newAvatarURL
	}
	user.UpdatedAt = time.Now()
	newUser, err := u.repo.Update(ctx, user)
	if err != nil {
		if newAvatarURL != nil {
			// DB更新に失敗した場合はアップロード済みのアバターを削除する（失敗しても更新エラーを優先して返す）
			_ = u.assetRepo.DeleteFile(ctx, *newAvatarURL)
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	if newAvatarURL != nil && oldAvatarURL != "" {
		// 旧アバターが自前のS3以外を指している場合もあるため、削除の失敗は無視する
		_ = u.assetRepo.DeleteFile(ctx, oldAvatarURL)
	}
	return newUser, nil
}
