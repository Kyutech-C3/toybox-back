package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
)

type CommentRepository interface {
	FindByWorkID(ctx context.Context, workID uuid.UUID) ([]*entity.Comment, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error)
	Create(ctx context.Context, comment *entity.Comment) (*entity.Comment, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}
