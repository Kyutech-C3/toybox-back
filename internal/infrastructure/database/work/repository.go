package work

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/types"
)

type WorkRepository struct {
	db *bun.DB
}

func NewWorkRepository(db *bun.DB) *WorkRepository {
	return &WorkRepository{
		db: db,
	}
}

func (r *WorkRepository) GetAll(ctx context.Context, limit, offset int, tagIDs []uuid.UUID) ([]*entity.Work, int, error) {
	var dtoWorks []*dto.Work

	countQuery := r.db.NewSelect().
		Model(&dtoWorks).
		Where("visibility IN (?)", bun.In([]types.Visibility{types.VisibilityPublic, types.VisibilityPrivate})).
		Where("EXISTS (SELECT 1 FROM asset WHERE asset.work_id = work.id)").
		Where("EXISTS (SELECT 1 FROM tagging WHERE tagging.work_id = work.id)")

	// タグIDsが指定されている場合はOR検索でフィルタリング
	if len(tagIDs) > 0 {
		countQuery = countQuery.Where("EXISTS (SELECT 1 FROM tagging WHERE tagging.work_id = work.id AND tagging.tag_id IN (?))", bun.In(tagIDs))
	}

	total, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, domainerrors.ErrFailedToGetAllWorksByLimitAndOffset
	}

	selectQuery := r.db.NewSelect().
		Model(&dtoWorks).
		Where("visibility IN (?)", bun.In([]types.Visibility{types.VisibilityPublic, types.VisibilityPrivate})).
		Where("EXISTS (SELECT 1 FROM asset WHERE asset.work_id = work.id)").
		Where("EXISTS (SELECT 1 FROM tagging WHERE tagging.work_id = work.id)")

	// タグIDsが指定されている場合はOR検索でフィルタリング
	if len(tagIDs) > 0 {
		selectQuery = selectQuery.Where("EXISTS (SELECT 1 FROM tagging WHERE tagging.work_id = work.id AND tagging.tag_id IN (?))", bun.In(tagIDs))
	}

	err = selectQuery.
		Relation("Assets").
		Relation("URLs").
		Relation("Tags").
		Relation("Collaborators").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, domainerrors.ErrWorkNotFound
		}
		return nil, 0, domainerrors.ErrFailedToGetAllWorksByLimitAndOffset
	}

	entityWorks := make([]*entity.Work, len(dtoWorks))
	for i, dtoWork := range dtoWorks {
		entityWorks[i] = dtoWork.ToWorkEntity()
	}

	return entityWorks, total, nil
}

func (r *WorkRepository) GetAllPublic(ctx context.Context, limit, offset int, tagIDs []uuid.UUID) ([]*entity.Work, int, error) {
	var dtoWorks []*dto.Work

	countQuery := r.db.NewSelect().
		Model(&dtoWorks).
		Where("visibility IN (?)", bun.In([]types.Visibility{types.VisibilityPublic})).
		Where("EXISTS (SELECT 1 FROM asset WHERE asset.work_id = work.id)").
		Where("EXISTS (SELECT 1 FROM tagging WHERE tagging.work_id = work.id)")

	// タグIDsが指定されている場合はOR検索でフィルタリング
	if len(tagIDs) > 0 {
		countQuery = countQuery.Where("EXISTS (SELECT 1 FROM tagging WHERE tagging.work_id = work.id AND tagging.tag_id IN (?))", bun.In(tagIDs))
	}

	total, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, domainerrors.ErrFailedToGetAllWorksByLimitAndOffset
	}

	selectQuery := r.db.NewSelect().
		Model(&dtoWorks).
		Where("visibility IN (?)", bun.In([]types.Visibility{types.VisibilityPublic})).
		Where("EXISTS (SELECT 1 FROM asset WHERE asset.work_id = work.id)").
		Where("EXISTS (SELECT 1 FROM tagging WHERE tagging.work_id = work.id)")

	// タグIDsが指定されている場合はOR検索でフィルタリング
	if len(tagIDs) > 0 {
		selectQuery = selectQuery.Where("EXISTS (SELECT 1 FROM tagging WHERE tagging.work_id = work.id AND tagging.tag_id IN (?))", bun.In(tagIDs))
	}

	err = selectQuery.
		Relation("Assets").
		Relation("URLs").
		Relation("Tags").
		Relation("Collaborators").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, domainerrors.ErrWorkNotFound
		}
		return nil, 0, domainerrors.ErrFailedToGetAllWorksByLimitAndOffset
	}

	entityWorks := make([]*entity.Work, len(dtoWorks))
	for i, dtoWork := range dtoWorks {
		entityWorks[i] = dtoWork.ToWorkEntity()
	}

	return entityWorks, total, nil
}

func (r *WorkRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Work, error) {
	var dtoWork dto.Work
	err := r.db.NewSelect().
		Model(&dtoWork).
		Relation("Assets").
		Relation("Tags").
		Relation("URLs").
		Relation("Collaborators").
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrWorkNotFound
		}
		return nil, domainerrors.ErrFailedToGetWorkById
	}

	return dtoWork.ToWorkEntity(), nil
}

func (r *WorkRepository) GetByUserID(ctx context.Context, userID uuid.UUID, public bool) ([]*entity.Work, error) {
	var dtoWorks []*dto.Work
	if public {
		err := r.db.NewSelect().
			Model(&dtoWorks).
			Where("user_id = ?", userID).
			Where("visibility IN (?)", bun.In([]types.Visibility{types.VisibilityPublic})).
			Where("EXISTS (SELECT 1 FROM asset WHERE asset.work_id = work.id)").
			Where("EXISTS (SELECT 1 FROM tagging WHERE tagging.work_id = work.id)").
			Relation("Assets").
			Relation("URLs").
			Relation("Tags").
			Relation("Collaborators").
			Scan(ctx)
		if err != nil {
			return nil, domainerrors.ErrFailedToGetWorksByUserID
		}
	} else {
		err := r.db.NewSelect().
			Model(&dtoWorks).
			Where("user_id = ?", userID).
			Where("visibility IN (?)", bun.In([]types.Visibility{types.VisibilityPublic, types.VisibilityPrivate})).
			Where("EXISTS (SELECT 1 FROM asset WHERE asset.work_id = work.id)").
			Where("EXISTS (SELECT 1 FROM tagging WHERE tagging.work_id = work.id)").
			Relation("Assets").
			Relation("URLs").
			Relation("Tags").
			Relation("Collaborators").
			Scan(ctx)
		if err != nil {
			return nil, domainerrors.ErrFailedToGetWorksByUserID
		}
	}
	entityWorks := make([]*entity.Work, len(dtoWorks))
	for i, dtoWork := range dtoWorks {
		entityWorks[i] = dtoWork.ToWorkEntity()
	}
	return entityWorks, nil
}

func (r *WorkRepository) ExistsById(ctx context.Context, id uuid.UUID) (bool, error) {
	var dtoWork dto.Work
	exists, err := r.db.NewSelect().
		Model(&dtoWork).
		Where("id = ?", id).
		Exists(ctx)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *WorkRepository) Create(ctx context.Context, work *entity.Work) (*entity.Work, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, domainerrors.ErrFailedToBeginTransaction
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	dtoWork := dto.ToWorkDTO(work)

	_, err = tx.NewInsert().Model(dtoWork).Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToCreateWork
	}
	thumbnail := &dto.Thumbnail{
		WorkID:  dtoWork.ID,
		AssetID: dtoWork.ThumbnailAssetID,
	}

	_, err = tx.NewInsert().Model(thumbnail).Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToCreateThumbnail
	}

	for _, asset := range dtoWork.Assets {
		_, err = tx.NewUpdate().Model(asset).Set("work_id = ?", dtoWork.ID).Where("id = ?", asset.ID).Exec(ctx)
		if err != nil {
			return nil, domainerrors.ErrFailedToCreateAsset
		}
	}

	_, err = tx.NewUpdate().Model(&dto.Asset{}).Set("work_id = ?", dtoWork.ID).Where("id = ?", dtoWork.ThumbnailAssetID).Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToCreateAsset
	}

	if len(dtoWork.URLs) > 0 {
		_, err = tx.NewInsert().Model(&dtoWork.URLs).Exec(ctx)
		if err != nil {
			return nil, domainerrors.ErrFailedToCreateURL
		}
	}

	if len(dtoWork.TagIDs) > 0 {
		taggings := make([]*dto.Tagging, len(dtoWork.TagIDs))
		for i, tagID := range dtoWork.TagIDs {
			taggings[i] = &dto.Tagging{
				WorkID: dtoWork.ID,
				TagID:  tagID,
			}
		}
		_, err = tx.NewInsert().Model(&taggings).Exec(ctx)
		if err != nil {
			return nil, domainerrors.ErrFailedToCreateTagging
		}
	}

	if len(dtoWork.Collaborators) > 0 {
		collaborators := make([]*dto.Collaborator, len(dtoWork.Collaborators))
		for i, collaborator := range dtoWork.Collaborators {
			collaborators[i] = &dto.Collaborator{
				WorkID: dtoWork.ID,
				UserID: collaborator.ID,
			}
		}
		_, err = tx.NewInsert().Model(&collaborators).Exec(ctx)
		if err != nil {
			return nil, domainerrors.ErrFailedToCreateWork
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, domainerrors.ErrFailedToCommitTransaction
	}

	return dtoWork.ToWorkEntity(), nil
}

func (r *WorkRepository) Update(ctx context.Context, work *entity.Work) (*entity.Work, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, domainerrors.ErrFailedToBeginTransaction
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	dtoWork := dto.ToWorkDTO(work)
	_, err = tx.NewUpdate().Model(dtoWork).WherePK().Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToUpdateWork
	}

	_, err = tx.NewDelete().Model(&dto.Tagging{}).Where("work_id = ?", work.ID).Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToUpdateWork
	}
	_, err = tx.NewDelete().Model(&dto.URLInfo{}).Where("work_id = ?", work.ID).Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToUpdateWork
	}
	_, err = tx.NewDelete().Model(&dto.Thumbnail{}).Where("work_id = ?", work.ID).Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToUpdateWork
	}
	_, err = tx.NewDelete().Model(&dto.Collaborator{}).Where("work_id = ?", work.ID).Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToUpdateWork
	}

	thumbnail := &dto.Thumbnail{
		WorkID:  dtoWork.ID,
		AssetID: dtoWork.ThumbnailAssetID,
	}
	_, err = tx.NewInsert().Model(thumbnail).Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToUpdateWork
	}

	if len(dtoWork.URLs) > 0 {
		_, err = tx.NewInsert().Model(&dtoWork.URLs).Exec(ctx)
		if err != nil {
			return nil, domainerrors.ErrFailedToUpdateWork
		}
	}

	if len(dtoWork.TagIDs) > 0 {
		taggings := make([]*dto.Tagging, len(dtoWork.TagIDs))
		for i, tagID := range dtoWork.TagIDs {
			taggings[i] = &dto.Tagging{
				WorkID: dtoWork.ID,
				TagID:  tagID,
			}
		}
		_, err = tx.NewInsert().Model(&taggings).Exec(ctx)
		if err != nil {
			return nil, domainerrors.ErrFailedToUpdateWork
		}
	}

	if len(dtoWork.Collaborators) > 0 {
		collaborators := make([]*dto.Collaborator, len(dtoWork.Collaborators))
		for i, collaborator := range dtoWork.Collaborators {
			collaborators[i] = &dto.Collaborator{
				WorkID: dtoWork.ID,
				UserID: collaborator.ID,
			}
		}
		_, err = tx.NewInsert().Model(&collaborators).Exec(ctx)
		if err != nil {
			return nil, domainerrors.ErrFailedToUpdateWork
		}
	}

	if len(dtoWork.Assets) > 0 {
		for _, asset := range dtoWork.Assets {
			_, err = tx.NewUpdate().Model(asset).Set("work_id = ?", dtoWork.ID).Where("id = ?", asset.ID).Exec(ctx)
			if err != nil {
				return nil, domainerrors.ErrFailedToUpdateWork
			}
		}
	}
	_, err = tx.NewUpdate().Model(&dto.Asset{}).Set("work_id = ?", dtoWork.ID).Where("id = ?", dtoWork.ThumbnailAssetID).Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToUpdateWork
	}

	err = tx.Commit()
	if err != nil {
		return nil, domainerrors.ErrFailedToCommitTransaction
	}

	return r.GetByID(ctx, work.ID)
}

func (r *WorkRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {

	work, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if work.UserID != userID {
		return domainerrors.ErrWorkNotOwnedByUser
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domainerrors.ErrFailedToBeginTransaction
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.NewDelete().Model(&dto.Tagging{}).Where("work_id = ?", id).Exec(ctx)
	if err != nil {
		return domainerrors.ErrFailedToDeleteWork
	}

	_, err = tx.NewDelete().Model(&dto.Thumbnail{}).Where("work_id = ?", id).Exec(ctx)
	if err != nil {
		return domainerrors.ErrFailedToDeleteWork
	}

	_, err = tx.NewDelete().Model(&dto.URLInfo{}).Where("work_id = ?", id).Exec(ctx)
	if err != nil {
		return domainerrors.ErrFailedToDeleteWork
	}

	_, err = tx.NewDelete().Model(&dto.Collaborator{}).Where("work_id = ?", id).Exec(ctx)
	if err != nil {
		return domainerrors.ErrFailedToDeleteWork
	}

	_, err = tx.NewDelete().Model(&dto.Work{}).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return domainerrors.ErrFailedToDeleteWork
	}

	_, err = tx.NewDelete().Model(&dto.Asset{}).Where("work_id = ?", id).Exec(ctx)
	if err != nil {
		return domainerrors.ErrFailedToDeleteWork
	}

	err = tx.Commit()
	if err != nil {
		return domainerrors.ErrFailedToCommitTransaction
	}

	return nil
}
