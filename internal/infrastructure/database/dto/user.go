package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
)

type User struct {
	bun.BaseModel `bun:"table:user"`

	ID            uuid.UUID `bun:"id,pk,default:gen_random_uuid()"`
	Name          string    `bun:"name,notnull"`
	Email         string    `bun:"email,notnull,unique"`
	DisplayName   string    `bun:"display_name,notnull"`
	DiscordUserID string    `bun:"discord_user_id"`
	Profile       string    `bun:"profile"`
	AvatarURL     string    `bun:"avatar_url"`
	XUsername     string    `bun:"x_username"`
	GithubID      string    `bun:"github_id"`
	CreatedAt     time.Time `bun:"created_at,notnull"`
	UpdatedAt     time.Time `bun:"updated_at,notnull"`
}

func (u *User) ToUserEntity() *entity.User {
	return &entity.User{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		DiscordUserID: u.DiscordUserID,
		Profile:       u.Profile,
		AvatarURL:     u.AvatarURL,
		XUsername:     u.XUsername,
		GithubID:      u.GithubID,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

func ToUserDTO(entity *entity.User) *User {
	return &User{
		ID:            entity.ID,
		Name:          entity.Name,
		Email:         entity.Email,
		DisplayName:   entity.DisplayName,
		DiscordUserID: entity.DiscordUserID,
		Profile:       entity.Profile,
		AvatarURL:     entity.AvatarURL,
		XUsername:     entity.XUsername,
		GithubID:      entity.GithubID,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
	}
}
