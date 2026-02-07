package entity

import (
	"time"

	"github.com/google/uuid"
)

type Work struct {
	ID               uuid.UUID
	Title            string
	Description      string
	DescriptionHTML  string
	UserID           uuid.UUID
	User             *User
	Visibility       string
	ThumbnailAssetID uuid.UUID
	ThumbnailURL     string
	Assets           []*Asset
	URLs             []*string
	TagIDs           []uuid.UUID
	Tags             []*Tag
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewWork(title string, description string, userID uuid.UUID, visibility string, thumbnailAssetID uuid.UUID, assets []*Asset, urls []*string, tagIDs []uuid.UUID, tags []*Tag) *Work {
	return &Work{
		ID:               uuid.New(),
		Title:            title,
		Description:      description,
		UserID:           userID,
		Visibility:       visibility,
		ThumbnailAssetID: thumbnailAssetID,
		Assets:           assets,
		URLs:             urls,
		TagIDs:           tagIDs,
		Tags:             tags,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}
