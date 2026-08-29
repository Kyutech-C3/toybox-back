package comment

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

type CommentRepository struct {
	db *bun.DB
}

func NewCommentRepository(db *bun.DB) *CommentRepository {
	return &CommentRepository{
		db: db,
	}
}

func (r *CommentRepository) FindByWorkID(ctx context.Context, workID uuid.UUID) ([]*entity.Comment, error) {
	var dtoComments []*dto.Comment
	err := r.db.NewSelect().
		Model(&dtoComments).
		Where("work_id = ?", workID).
		Relation("User").
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]*entity.Comment, 0), nil
		}
		return nil, domainerrors.ErrFailedToGetCommentsByWorkID
	}

	entityComments := make([]*entity.Comment, len(dtoComments))
	for i, dtoComment := range dtoComments {
		entityComments[i] = dtoComment.ToCommentEntity()
	}

	return entityComments, nil
}

func (r *CommentRepository) Create(ctx context.Context, comment *entity.Comment) (*entity.Comment, error) {
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

	dtoComment := dto.ToCommentDTO(comment)

	_, err = tx.NewInsert().Model(dtoComment).Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToCreateComment
	}

	if err := tx.Commit(); err != nil {
		return nil, domainerrors.ErrFailedToCommitTransaction
	}

	return dtoComment.ToCommentEntity(), nil
}

func (r *CommentRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error) {
	var dtoComment dto.Comment
	err := r.db.NewSelect().
		Model(&dtoComment).
		Where("comment.id = ?", id).
		Relation("User").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrCommentNotFound
		}
		return nil, domainerrors.ErrFailedToGetCommentById
	}

	return dtoComment.ToCommentEntity(), nil
}

func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	comment, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if comment.UserID != userID {
		return domainerrors.ErrCommentNotOwnedByUser
	}

	_, err = r.db.NewUpdate().Model(&dto.Comment{}).
		Set("status = ?", types.CommentStatusDeleted).
		Where("id = ?", id).Exec(ctx)
	if err != nil {
		return domainerrors.ErrFailedToDeleteComment
	}
	return nil
}
