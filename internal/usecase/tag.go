package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/domain/repository"
)

type ITagUseCase interface {
	Create(ctx context.Context, name string) (*entity.Tag, error)
	GetAll(ctx context.Context, authenticated bool) ([]*entity.Tag, error)
}

type tagUseCase struct {
	tagRepo repository.TagRepository
}

func NewTagUseCase(tagRepo repository.TagRepository) ITagUseCase {
	return &tagUseCase{
		tagRepo: tagRepo,
	}
}

func (uc *tagUseCase) Create(ctx context.Context, name string) (*entity.Tag, error) {
	if name == "" {
		return nil, domainerrors.ErrInvalidTagName
	}

	now := time.Now()
	tag := &entity.Tag{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	tag.NormalizeName()

	exists, err := uc.tagRepo.ExistsByName(ctx, tag.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check tag existence: %w", err)
	}
	if exists {
		return nil, domainerrors.ErrTagAlreadyExists
	}

	createdTag, err := uc.tagRepo.Create(ctx, tag)
	if err != nil {
		return nil, err
	}

	return createdTag, nil
}

func (uc *tagUseCase) GetAll(ctx context.Context, authenticated bool) ([]*entity.Tag, error) {
	tags, err := uc.tagRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	counts, err := uc.tagRepo.CountWorksByTag(ctx, authenticated)
	if err != nil {
		return nil, err
	}
	for _, tag := range tags {
		tag.WorkCount = counts[tag.ID]
	}

	return tags, nil
}
