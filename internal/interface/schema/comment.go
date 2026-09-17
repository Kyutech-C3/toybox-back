package schema

import (
	"time"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
)

type UserInCommentResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type CommentResponse struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Status    string                 `json:"status"`
	ReplyAt   string                 `json:"reply_at"`
	User      *UserInCommentResponse `json:"user"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

type CreateCommentRequest struct {
	Content string `json:"content" validate:"required,max=255"`
	ReplyAt string `json:"reply_at" validate:"omitempty,uuid"`
	UserID  string `json:"user_id" validate:"omitempty,uuid"`
}

type CreateCommentResponse struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	ReplyAt   string `json:"reply_at"`
	CreatedAt string `json:"created_at"`
}

func ToCommentResponse(comment *entity.Comment) *CommentResponse {
	if comment == nil {
		return nil
	}

	if comment.Status == "deleted" {
		return &CommentResponse{
			ID:        comment.ID.String(),
			Content:   "",
			Status:    comment.Status,
			ReplyAt:   comment.ReplyAt,
			User:      nil,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
			UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
		}
	}

	var user *UserInCommentResponse
	if comment.User != nil && comment.User.ID != uuid.Nil {
		user = &UserInCommentResponse{
			ID:          comment.User.ID.String(),
			DisplayName: comment.User.DisplayName,
			AvatarURL:   comment.User.AvatarURL,
		}
	}

	return &CommentResponse{
		ID:        comment.ID.String(),
		Content:   comment.Content,
		Status:    comment.Status,
		ReplyAt:   comment.ReplyAt,
		User:      user,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
	}
}

func ToCommentListResponse(comments []*entity.Comment) []*CommentResponse {
	res := make([]*CommentResponse, len(comments))
	for i, comment := range comments {
		res[i] = ToCommentResponse(comment)
	}
	return res
}

func ToCreateCommentResponse(comment *entity.Comment) *CreateCommentResponse {
	if comment == nil {
		return nil
	}
	return &CreateCommentResponse{
		ID:        comment.ID.String(),
		Content:   comment.Content,
		ReplyAt:   comment.ReplyAt,
		CreatedAt: comment.CreatedAt.String(),
	}
}
