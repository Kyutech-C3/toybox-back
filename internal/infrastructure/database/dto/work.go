package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/types"
)

type Work struct {
	bun.BaseModel `bun:"table:work"`

	ID               uuid.UUID        `bun:"id,pk"`
	Title            string           `bun:"title,notnull"`
	Description      string           `bun:"description,notnull"`
	Visibility       types.Visibility `bun:"visibility"`
	ThumbnailAssetID uuid.UUID        `bun:"-"`
	Assets           []*Asset         `bun:"rel:has-many,join:id=work_id"`
	URLs             []*URLInfo       `bun:"rel:has-many,join:id=work_id"`
	Tags             []*Tag           `bun:"m2m:tagging,join:Work=Tag"`
	TagIDs           []uuid.UUID      `bun:"-"`
	Collaborators    []*User          `bun:"m2m:collaborator"`
	UserID           uuid.UUID        `bun:"user_id,notnull"`
	CreatedAt        time.Time        `bun:"created_at,notnull"`
	UpdatedAt        time.Time        `bun:"updated_at,notnull"`
}

func (w *Work) ToWorkEntity() *entity.Work {
	assets := make([]*entity.Asset, len(w.Assets))
	for i, asset := range w.Assets {
		assets[i] = asset.ToAssetEntity()
	}

	urls := make([]*string, len(w.URLs))
	for i, url := range w.URLs {
		urls[i] = url.ToURLInfoEntity()
	}

	entityTags := make([]*entity.Tag, len(w.Tags))
	tagIDs := make([]uuid.UUID, len(w.Tags))
	for i, tag := range w.Tags {
		entityTags[i] = tag.ToTagEntity()
		tagIDs[i] = tag.ID
	}

	collaborators := make([]*entity.User, len(w.Collaborators))
	for i, collaborator := range w.Collaborators {
		collaborators[i] = collaborator.ToUserEntity()
	}

	return &entity.Work{
		ID:            w.ID,
		Title:         w.Title,
		Description:   w.Description,
		UserID:        w.UserID,
		Visibility:    string(w.Visibility),
		Assets:        assets,
		URLs:          urls,
		TagIDs:        tagIDs,
		Tags:          entityTags,
		Collaborators: collaborators,
		CreatedAt:     w.CreatedAt,
		UpdatedAt:     w.UpdatedAt,
	}
}

func ToWorkDTO(entity *entity.Work) *Work {
	assets := make([]*Asset, len(entity.Assets))
	for i, asset := range entity.Assets {
		assets[i] = ToAssetDTO(asset)
	}

	urls := make([]*URLInfo, len(entity.URLs))
	for i, url := range entity.URLs {
		urls[i] = ToURLInfoDTO(entity.ID, *url, entity.UserID)
	}

	tags := make([]*Tag, len(entity.Tags))
	for i, tag := range entity.Tags {
		tags[i] = ToTagDTO(tag)
	}

	collaborators := make([]*User, len(entity.Collaborators))
	for i, collaborator := range entity.Collaborators {
		collaborators[i] = ToUserDTO(collaborator)
	}

	return &Work{
		ID:               entity.ID,
		Title:            entity.Title,
		Description:      entity.Description,
		UserID:           entity.UserID,
		ThumbnailAssetID: entity.ThumbnailAssetID,
		Visibility:       types.Visibility(entity.Visibility),
		Assets:           assets,
		URLs:             urls,
		Tags:             tags,
		TagIDs:           entity.TagIDs,
		Collaborators:    collaborators,
		CreatedAt:        entity.CreatedAt,
		UpdatedAt:        entity.UpdatedAt,
	}
}
