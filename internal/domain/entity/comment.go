package entity

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID
	Content   string
	WorkID    uuid.UUID
	UserID    uuid.UUID
	ReplyAt   string
	User      *User
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewComment(content string, workID uuid.UUID, userID uuid.UUID, replyAt string) *Comment {
	return &Comment{
		ID:        uuid.New(),
		Content:   content,
		WorkID:    workID,
		UserID:    userID,
		ReplyAt:   replyAt,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
