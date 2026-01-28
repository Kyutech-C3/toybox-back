package schema

import (
	"time"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
)

type GetWorkOutput struct {
	ID            uuid.UUID              `json:"id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	UserID        uuid.UUID              `json:"user_id"`
	Visibility    string                 `json:"visibility"`
	Assets        []AssetResponse        `json:"assets"`
	Tags          []TagResponse          `json:"tags"`
	Collaborators []CollaboratorResponse `json:"collaborators"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

type CreateWorkInput struct {
	Title            string      `json:"title" validate:"required,max=100"`
	Description      string      `json:"description" validate:"required"`
	Visibility       string      `json:"visibility" validate:"required,oneof=public private draft"`
	ThumbnailAssetID uuid.UUID   `json:"thumbnail_asset_id" validate:"required,uuid"`
	AssetIDs         []uuid.UUID `json:"asset_ids" validate:"required,dive,uuid"`
	URLs             []string    `json:"urls" validate:"required,dive,url"`
	TagIDs           []uuid.UUID `json:"tag_ids" validate:"required,dive,uuid"`
	CollaboratorIDs  []uuid.UUID `json:"collaborator_ids,omitempty" validate:"omitempty,dive,uuid"`
}

type CreateWorkOutput struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	UserID      uuid.UUID `json:"user_id"`
	Visibility  string    `json:"visibility"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

type UpdateWorkInput struct {
	Title            *string      `json:"title,omitempty" validate:"omitempty,max=100"`
	Description      *string      `json:"description,omitempty"`
	Visibility       *string      `json:"visibility,omitempty" validate:"omitempty,oneof=public private draft"`
	ThumbnailAssetID *uuid.UUID   `json:"thumbnail_asset_id,omitempty" validate:"omitempty,uuid"`
	AssetIDs         *[]uuid.UUID `json:"asset_ids,omitempty" validate:"omitempty,dive,uuid"`
	URLs             *[]string    `json:"urls,omitempty" validate:"omitempty,dive,url"`
	TagIDs           *[]uuid.UUID `json:"tag_ids,omitempty" validate:"omitempty,dive,uuid"`
	CollaboratorIDs  *[]uuid.UUID `json:"collaborator_ids,omitempty" validate:"omitempty,dive,uuid"`
}

type GetWorksQuery struct {
	Limit  *int   `query:"limit" validate:"omitempty,min=1,max=100"`
	Page   *int   `query:"page" validate:"omitempty,min=1"`
	TagIDs string `query:"tag_ids" validate:"omitempty"`
}

type WorkListResponse struct {
	Works      []GetWorkOutput `json:"works"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
}

type AssetResponse struct {
	ID        uuid.UUID `json:"id"`
	WorkID    uuid.UUID `json:"work_id"`
	AssetType string    `json:"asset_type"`
	UserID    uuid.UUID `json:"user_id"`
	Extension string    `json:"extension"`
	URL       string    `json:"url"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type TagResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type CollaboratorResponse struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
}

func ToWorkResponse(work *entity.Work) GetWorkOutput {
	if work == nil {
		return GetWorkOutput{}
	}
	return GetWorkOutput{
		ID:            work.ID,
		Title:         work.Title,
		Description:   work.Description,
		UserID:        work.UserID,
		Visibility:    work.Visibility,
		Assets:        ToAssetResponses(work.Assets),
		Tags:          ToTagResponses(work.Tags),
		Collaborators: ToCollaboratorResponses(work.Collaborators),
		CreatedAt:     work.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     work.UpdatedAt.Format(time.RFC3339),
	}
}

func ToCreateWorkOutput(work *entity.Work) CreateWorkOutput {
	if work == nil {
		return CreateWorkOutput{}
	}
	return CreateWorkOutput{
		ID:          work.ID,
		Title:       work.Title,
		Description: work.Description,
		UserID:      work.UserID,
		Visibility:  work.Visibility,
		CreatedAt:   work.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   work.UpdatedAt.Format(time.RFC3339),
	}
}

func ToAssetResponse(asset *entity.Asset) AssetResponse {
	if asset == nil {
		return AssetResponse{}
	}

	return AssetResponse{
		ID:        asset.ID,
		WorkID:    asset.WorkID,
		AssetType: asset.AssetType,
		UserID:    asset.UserID,
		Extension: asset.Extension,
		URL:       asset.URL,
		CreatedAt: asset.CreatedAt.Format(time.RFC3339),
		UpdatedAt: asset.UpdatedAt.Format(time.RFC3339),
	}
}

func ToAssetResponses(assets []*entity.Asset) []AssetResponse {
	if len(assets) == 0 {
		return []AssetResponse{}
	}

	res := make([]AssetResponse, 0, len(assets))
	for _, asset := range assets {
		res = append(res, ToAssetResponse(asset))
	}
	return res
}

func ToTagResponse(tag *entity.Tag) TagResponse {
	if tag == nil {
		return TagResponse{}
	}

	return TagResponse{
		ID:   tag.ID,
		Name: tag.Name,
	}
}

func ToTagResponses(tags []*entity.Tag) []TagResponse {
	if len(tags) == 0 {
		return []TagResponse{}
	}

	res := make([]TagResponse, 0, len(tags))
	for _, tag := range tags {
		res = append(res, ToTagResponse(tag))
	}
	return res
}

func ToCollaboratorResponse(user *entity.User) CollaboratorResponse {
	if user == nil {
		return CollaboratorResponse{}
	}

	return CollaboratorResponse{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
	}
}

func ToCollaboratorResponses(users []*entity.User) []CollaboratorResponse {
	if len(users) == 0 {
		return []CollaboratorResponse{}
	}

	res := make([]CollaboratorResponse, 0, len(users))
	for _, user := range users {
		res = append(res, ToCollaboratorResponse(user))
	}
	return res
}
func ToWorkListResponse(works []*entity.Work) WorkListResponse {
	if len(works) == 0 {
		return WorkListResponse{}
	}
	workResponses := make([]GetWorkOutput, 0, len(works))
	for _, work := range works {
		workResponses = append(workResponses, ToWorkResponse(work))
	}
	return WorkListResponse{
		Works:      workResponses,
		TotalCount: len(works),
		Page:       1,
		Limit:      20,
	}
}
