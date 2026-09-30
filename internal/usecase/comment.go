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

type ICommentUsecase interface {
	GetCommentsByWorkID(ctx context.Context, workID uuid.UUID, userID uuid.UUID) ([]*entity.Comment, error)
	CreateComment(ctx context.Context, content string, workID, userID uuid.UUID, replyAt string) (*entity.Comment, error)
}

type commentUsecase struct {
	commentRepo repository.CommentRepository
	workRepo    repository.WorkRepository
	timeout     time.Duration
}

func NewCommentUsecase(commentRepo repository.CommentRepository, workRepo repository.WorkRepository, timeout time.Duration) ICommentUsecase {
	return &commentUsecase{
		commentRepo: commentRepo,
		workRepo:    workRepo,
		timeout:     time.Second * 30,
	}
}

func (uc *commentUsecase) GetCommentsByWorkID(ctx context.Context, workID uuid.UUID, userID uuid.UUID) ([]*entity.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	work, err := uc.workRepo.GetByID(ctx, workID)
	if err != nil {
		return nil, fmt.Errorf("failed to get work by ID %s: %w", workID.String(), err)
	}
	if !canViewWork(work, userID) {
		return nil, domainerrors.ErrWorkNotViewable
	}

	comments, err := uc.commentRepo.FindByWorkID(ctx, workID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by work ID %s: %w", workID.String(), err)
	}

	return comments, nil
}

func (uc *commentUsecase) CreateComment(ctx context.Context, content string, workID, userID uuid.UUID, replyAt string) (*entity.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if workID == uuid.Nil {
		return nil, fmt.Errorf("work ID is required")
	}
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	work, err := uc.workRepo.GetByID(ctx, workID)
	if err != nil {
		return nil, fmt.Errorf("failed to get work by ID %s: %w", workID.String(), err)
	}
	if !canViewWork(work, userID) {
		return nil, domainerrors.ErrWorkNotViewable
	}

	// replyAtがある場合は返信先にコメントが存在するか確認
	if replyAt != "" {
		replyID, err := uuid.Parse(replyAt)
		if err != nil {
			return nil, fmt.Errorf("invalid reply_at format: %w", err)
		}
		replyComment, err := uc.commentRepo.FindByID(ctx, replyID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate reply target comment %s: %w", replyAt, err)
		}
		if replyComment.WorkID != workID {
			return nil, domainerrors.ErrInvalidReplyAt
		}
	}
	comment := entity.NewComment(content, workID, userID, replyAt)

	createdComment, err := uc.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return createdComment, nil
}

func canViewWork(work *entity.Work, userID uuid.UUID) bool {
	switch work.Visibility {
	case "public":
		return true
	case "private":
		return userID != uuid.Nil
	default:
		return false
	}
}
