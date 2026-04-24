package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/domain/repository"
)

type ITagUseCase interface {
	Create(ctx context.Context, name string) (*entity.Tag, error)
	GetAll(ctx context.Context) ([]*entity.Tag, error)
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
	createdTag, err := uc.tagRepo.Create(ctx, tag)
	if err != nil {
		return nil, err
	}

	return createdTag, nil
}

func (uc *tagUseCase) GetAll(ctx context.Context) ([]*entity.Tag, error) {
	tags, err := uc.tagRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return tags, nil
}
