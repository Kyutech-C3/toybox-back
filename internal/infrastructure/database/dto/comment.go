package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
)

type Comment struct {
	bun.BaseModel `bun:"table:comment"`
	ID            uuid.UUID        `json:"id" bun:"id,pk"`
	Content       string           `json:"content" bun:"content,notnull"`
	WorkID        uuid.UUID        `json:"work_id" bun:"work_id,notnull"`
	UserID        uuid.UUID        `json:"user_id" bun:"user_id"`
	ReplyAt       string           `json:"reply_at" bun:"reply_at"`
	User          *User            `bun:"rel:belongs-to,join:user_id=id"`
	CreatedAt     time.Time        `json:"created_at" bun:"created_at,notnull"`
	UpdatedAt     time.Time        `json:"updated_at" bun:"updated_at,notnull"`
}

func (c *Comment) ToCommentEntity() *entity.Comment {
	var user *entity.User
	if c.User != nil && c.User.ID != uuid.Nil {
		user = c.User.ToUserEntity()
	}

	return &entity.Comment{
		ID:         c.ID,
		Content:    c.Content,
		WorkID:     c.WorkID,
		UserID:     c.UserID,
		ReplyAt:    c.ReplyAt,
		User:       user,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

func ToCommentDTO(e *entity.Comment) *Comment {
	return &Comment{
		ID:        e.ID,
		Content:   e.Content,
		WorkID:    e.WorkID,
		UserID:    e.UserID,
		ReplyAt:   e.ReplyAt,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}
