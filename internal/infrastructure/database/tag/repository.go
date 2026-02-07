package tag

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"
)

type TagRepository struct {
	db *bun.DB
}

func NewTagRepository(db *bun.DB) *TagRepository {
	return &TagRepository{
		db: db,
	}
}

func (r *TagRepository) ExistAll(ctx context.Context, ids []uuid.UUID) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}

	count, err := r.db.NewSelect().
		Table("tag").
		Where("id IN (?)", bun.In(ids)).
		Count(ctx)
	if err != nil {
		return false, err
	}

	return count == len(ids), nil
}

func (r *TagRepository) FindAllByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Tag, error) {
	var dtoTags []*dto.Tag
	err := r.db.NewSelect().
		Model(&dtoTags).
		Where("id IN (?)", bun.In(ids)).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	entityTags := make([]*entity.Tag, len(dtoTags))
	for i, dtoTag := range dtoTags {
		entityTags[i] = dtoTag.ToTagEntity()
	}

	return entityTags, nil
}

func (r *TagRepository) Create(ctx context.Context, tag *entity.Tag) (*entity.Tag, error) {
	dtoTag := dto.ToTagDTO(tag)
	_, err := r.db.NewInsert().
		Model(dtoTag).
		Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToCreateTag
	}
	return dtoTag.ToTagEntity(), nil
}

func (r *TagRepository) FindAll(ctx context.Context) ([]*entity.Tag, error) {
	var dtoTags []*dto.Tag
	err := r.db.NewSelect().
		Model(&dtoTags).
		Order("name ASC").
		Scan(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToGetAllTags
	}

	entityTags := make([]*entity.Tag, len(dtoTags))
	for i, dtoTag := range dtoTags {
		entityTags[i] = dtoTag.ToTagEntity()
	}

	return entityTags, nil
}
