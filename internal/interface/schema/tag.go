package schema

import (
	"time"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
)

// Request
type CreateTagInput struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}

// Response
type TagDetailResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TagListResponse struct {
	Tags []TagDetailResponse `json:"tags"`
}

// Converter functions
func ToTagDetailResponse(tag *entity.Tag) TagDetailResponse {
	return TagDetailResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}
}

func ToTagListResponse(tags []*entity.Tag) TagListResponse {
	response := make([]TagDetailResponse, len(tags))
	for i, tag := range tags {
		response[i] = ToTagDetailResponse(tag)
	}
	return TagListResponse{Tags: response}
}

