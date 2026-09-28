package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID
	Name          string
	Email         string
	DisplayName   string
	DiscordUserID string
	Profile       string
	AvatarURL     string
	XUsername     string
	GithubID      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewUser(name string, email string, displayName string, discordUserID string, avatarURL string) *User {
	return &User{
		ID:            uuid.New(),
		Name:          name,
		Email:         email,
		DisplayName:   displayName,
		DiscordUserID: discordUserID,
		AvatarURL:     avatarURL,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
