package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	"github.com/simesaba80/toybox-back/internal/domain/repository"
)

type IUserUseCase interface {
	GetAllUser(ctx context.Context) ([]*entity.User, error)
	GetByUserID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, displayName *string, profile *string, twitterID *string, githubID *string) (*entity.User, error)
}

type userUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) IUserUseCase {
	return &userUseCase{
		repo: repo,
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

func (u *userUseCase) UpdateUser(ctx context.Context, userID uuid.UUID, displayName *string, profile *string, twitterID *string, githubID *string) (*entity.User, error) {
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
	if twitterID != nil {
		user.TwitterID = *twitterID
	}
	if githubID != nil {
		user.GithubID = *githubID
	}
	user.UpdatedAt = time.Now()
	newUser, err := u.repo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return newUser, nil
}
