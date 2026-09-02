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
	GetCommentsByWorkID(ctx context.Context, workID uuid.UUID) ([]*entity.Comment, error)
	CreateComment(ctx context.Context, content string, workID, userID uuid.UUID, replyAt string) (*entity.Comment, error)
	DeleteComment(ctx context.Context, id, userID uuid.UUID) error
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

func (uc *commentUsecase) GetCommentsByWorkID(ctx context.Context, workID uuid.UUID) ([]*entity.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	comments, err := uc.commentRepo.FindByWorkID(ctx, workID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by work ID %s: %w", workID.String(), err)
	}

	return filterVisibleComments(comments), nil
}

// 削除済みかつ表示可能な子孫を持たないコメントを一覧から除外します。
func filterVisibleComments(comments []*entity.Comment) []*entity.Comment {
	childrenOf := make(map[string][]*entity.Comment)
	for _, comment := range comments {
		if comment.ReplyAt != "" {
			childrenOf[comment.ReplyAt] = append(childrenOf[comment.ReplyAt], comment)
		}
	}

	visible := make(map[string]bool)
	res := make([]*entity.Comment, 0, len(comments))
	for _, comment := range comments {
		if !isCommentVisible(comment, childrenOf, visible) {
			continue
		}
		res = append(res, comment)
	}
	return res
}

// コメントが削除済みでも、表示可能な子孫を持つ場合はtrueを返します。
// visible は同じ一覧内での再帰呼び出し間で結果を使い回すためのメモ化キャッシュです
func isCommentVisible(comment *entity.Comment, childrenOf map[string][]*entity.Comment, visible map[string]bool) bool {
	if v, ok := visible[comment.ID.String()]; ok {
		return v
	}
	result := comment.Status != "deleted"
	if !result {
		for _, child := range childrenOf[comment.ID.String()] {
			if isCommentVisible(child, childrenOf, visible) {
				result = true
				break
			}
		}
	}
	visible[comment.ID.String()] = result
	return result
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

	// Workの存在確認
	exists, err := uc.workRepo.ExistsById(ctx, workID)
	if err != nil {
		return nil, fmt.Errorf("failed to check work existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("work not found: %s", workID.String())
	}

	// replyAtがある場合は返信先にコメントが存在するか確認
	if replyAt != "" {
		replyID, err := uuid.Parse(replyAt)
		if err != nil {
			return nil, fmt.Errorf("invalid reply_at format: %w", err)
		}
		_, err = uc.commentRepo.FindByID(ctx, replyID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate reply target comment %s: %w", replyAt, err)
		}
	}
	comment := entity.NewComment(content, workID, userID, replyAt)

	createdComment, err := uc.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return createdComment, nil
}

func (uc *commentUsecase) DeleteComment(ctx context.Context, id, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	comment, err := uc.commentRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get comment by ID %s: %w", id.String(), err)
	}
	if comment.Status == "deleted" {
		return domainerrors.ErrCommentNotFound
	}
	if comment.UserID != userID {
		return domainerrors.ErrCommentNotOwnedByUser
	}

	if err := uc.commentRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("failed to delete comment %s: %w", id.String(), err)
	}

	return nil
}
