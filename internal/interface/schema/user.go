package schema

import (
	"time"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
)

type GetUserOutput struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Profile     string    `json:"profile"`
	AvatarURL   string    `json:"avatar_url"`
	TwitterID   string    `json:"twitter_id"`
	GithubID    string    `json:"github_id"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

type GetIconAndURLResponse struct {
	DisplayName string `json:"display_name"`
	IconURL     string `json:"icon_url"`
}

type UserListResponse struct {
	Users []GetUserOutput `json:"users"`
}

type UpdateUserInput struct {
	Email       string `json:"email" validate:"required,email"`
	DisplayName string `json:"display_name" validate:"required,min=1,max=32"`
	Profile     string `json:"profile" validate:"omitempty,max=500"`
	TwitterID   string `json:"twitter_id" validate:"omitempty"`
	GithubID    string `json:"github_id" validate:"omitempty"`
}

func ToUserResponse(user *entity.User) GetUserOutput {
	if user == nil {
		return GetUserOutput{}
	}
	return GetUserOutput{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Profile:     user.Profile,
		AvatarURL:   user.AvatarURL,
		TwitterID:   user.TwitterID,
		GithubID:    user.GithubID,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
	}
}

func ToIconAndURLResponse(user *entity.User) GetIconAndURLResponse {
	if user == nil {
		return GetIconAndURLResponse{}
	}
	return GetIconAndURLResponse{
		DisplayName: user.DisplayName,
		IconURL:     user.AvatarURL,
	}
}
