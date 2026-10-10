package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
)

type DiscordWebhookData struct {
	Username  string                `json:"username"`
	AvatarURL string                `json:"avatar_url"`
	Embeds    []DiscordWebhookEmbed `json:"embeds"`
}

type DiscordWebhookEmbed struct {
	Title       string                `json:"title"`
	URL         string                `json:"url"`
	Description string                `json:"description"`
	Color       int                   `json:"color"`
	Image       []DiscordWebhookImage `json:"image"`
}

type DiscordWebhookImage struct {
	URL string `json:"url"`
}

func NewDiscordWebhookData(work *entity.Work) DiscordWebhookData {
	var description string
	if utf8.RuneCountInString(work.Description) > 100 {
		description = string([]rune(work.Description)[:100]) + "..."
	} else {
		description = work.Description
	}
	return DiscordWebhookData{
		Username:  work.User.Name,
		AvatarURL: work.User.AvatarURL,
		Embeds: []DiscordWebhookEmbed{
			{
				Title:       work.Title,
				URL:         config.FRONTEND_URL[0] + "/works/" + work.ID.String(),
				Description: description,
				Color:       4063189,
				Image:       []DiscordWebhookImage{{URL: work.ThumbnailURL}},
			},
		},
	}
}

type WebhookNotifierRepository struct {
	webhookURL string
}

func NewWebhookNotifierRepository() *WebhookNotifierRepository {
	return &WebhookNotifierRepository{
		webhookURL: config.WEBHOOK_URL,
	}
}

func (r *WebhookNotifierRepository) WebhookNotify(ctx context.Context, work *entity.Work) {
	discordWebhookData := NewDiscordWebhookData(work)
	jsonData, err := json.Marshal(discordWebhookData)
	if err != nil {
		log.Println("failed to marshal discord webhook data: %w", err)
		return
	}

	body := bytes.NewBuffer(jsonData)
	req, err := http.NewRequestWithContext(ctx, "POST", r.webhookURL, body)
	if err != nil {
		log.Println("failed to create request: %w", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		log.Println("failed to send webhook: %w", err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		log.Println("failed to send webhook: %w", err)
		return
	}

	return
}
