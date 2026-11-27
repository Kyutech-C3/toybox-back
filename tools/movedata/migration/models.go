package migration

import (
	"time"

	"github.com/uptrace/bun"
)

type OldUser struct {
	bun.BaseModel       `bun:"table:user"`
	ID                  string    `bun:"id,pk"`
	Name                string    `bun:"name"`
	Email               string    `bun:"email"`
	PasswordHash        *string   `bun:"password_hash"`
	DisplayName         string    `bun:"display_name"`
	DiscordToken        *string   `bun:"discord_token"`
	DiscordRefreshToken *string   `bun:"discord_refresh_token"`
	DiscordUserID       *string   `bun:"discord_user_id"`
	Profile             *string   `bun:"profile"`
	AvatarURL           *string   `bun:"avatar_url"`
	TwitterID           *string   `bun:"twitter_id"`
	GithubID            *string   `bun:"github_id"`
	CreatedAt           time.Time `bun:"created_at"`
	UpdatedAt           time.Time `bun:"updated_at"`
}

type OldWork struct {
	bun.BaseModel   `bun:"table:works"`
	ID              string    `bun:"id,pk"`
	Title           string    `bun:"title"`
	Description     string    `bun:"description"`
	DescriptionHTML string    `bun:"description_html"`
	UserID          *string   `bun:"user_id"`
	Visibility      string    `bun:"visibility"`
	CreatedAt       time.Time `bun:"created_at"`
	UpdatedAt       time.Time `bun:"updated_at"`
}

type OldFavorite struct {
	bun.BaseModel `bun:"table:favorite"`
	WorkID        string    `bun:"work_id,pk"`
	UserID        string    `bun:"user_id,pk"`
	CreatedAt     time.Time `bun:"created_at"`
}

type OldAsset struct {
	bun.BaseModel `bun:"table:asset"`
	ID            string    `bun:"id,pk"`
	WorkID        *string   `bun:"work_id"`
	AssetType     string    `bun:"asset_type"`
	UserID        string    `bun:"user_id"`
	Extension     string    `bun:"extension"`
	URL           string    `bun:"url"`
	CreatedAt     time.Time `bun:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at"`
}

type OldComment struct {
	bun.BaseModel `bun:"table:comment"`
	ID            string    `bun:"id,pk"`
	Content       string    `bun:"content"`
	WorkID        string    `bun:"work_id"`
	UserID        *string   `bun:"user_id"`
	ReplyAt       *string   `bun:"reply_at"`
	CreatedAt     time.Time `bun:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at"`
}

type OldURLInfo struct {
	bun.BaseModel `bun:"table:urlinfo"`
	ID            string    `bun:"id,pk"`
	WorkID        *string   `bun:"work_id"`
	URL           string    `bun:"url"`
	URLType       string    `bun:"url_type"`
	UserID        string    `bun:"user_id"`
	CreatedAt     time.Time `bun:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at"`
}

type OldTag struct {
	bun.BaseModel `bun:"table:tags"`
	ID            string    `bun:"id,pk"`
	Name          string    `bun:"name"`
	Color         string    `bun:"color"`
	CreatedAt     time.Time `bun:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at"`
}

type OldTagging struct {
	bun.BaseModel `bun:"table:taggings"`
	WorkID        string `bun:"work_id,pk"`
	TagID         string `bun:"tag_id,pk"`
}

type OldThumbnail struct {
	bun.BaseModel `bun:"table:thumbnails"`
	WorkID        string `bun:"work_id,pk"`
	AssetID       string `bun:"asset_id,pk"`
}
