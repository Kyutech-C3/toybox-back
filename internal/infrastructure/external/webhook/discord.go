package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
)

const (
	toyboxLogoURL      = "https://toybox.compositecomputer.club/_nuxt/img/ToyBoxlogo.21166b5.png"
	discordEmbedColor  = 4063189
	maxDescriptionRune = 100
)

type DiscordWorkNotifier struct {
	client      *http.Client
	webhookURL  string
	frontendURL string
}

type discordWebhookPayload struct {
	Username  string         `json:"username"`
	AvatarURL string         `json:"avatar_url"`
	Embeds    []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Title       string       `json:"title"`
	URL         string       `json:"url"`
	Description string       `json:"description"`
	Thumbnail   discordImage `json:"thumbnail"`
	Color       int          `json:"color"`
	Image       discordImage `json:"image"`
}

type discordImage struct {
	URL string `json:"url"`
}

func NewDiscordWorkNotifier() *DiscordWorkNotifier {
	frontendURL := ""
	if len(config.FRONTEND_URL) > 0 {
		frontendURL = config.FRONTEND_URL[0]
	}

	return newDiscordWorkNotifier(
		&http.Client{Timeout: 10 * time.Second},
		config.DISCORD_WEBHOOK_URL,
		frontendURL,
	)
}

func newDiscordWorkNotifier(client *http.Client, webhookURL, frontendURL string) *DiscordWorkNotifier {
	return &DiscordWorkNotifier{
		client:      client,
		webhookURL:  webhookURL,
		frontendURL: strings.TrimRight(frontendURL, "/"),
	}
}

func (n *DiscordWorkNotifier) NotifyCreated(ctx context.Context, work *entity.Work) error {
	if n.webhookURL == "" {
		return errors.New("DISCORD_WEBHOOK_URL is not configured")
	}
	if work == nil || work.User == nil {
		return errors.New("work or work user is nil")
	}

	payload := discordWebhookPayload{
		Username:  work.User.Name,
		AvatarURL: work.User.AvatarURL,
		Embeds: []discordEmbed{
			{
				Title:       work.Title,
				URL:         fmt.Sprintf("%s/works/%s", n.frontendURL, work.ID),
				Description: truncateDescription(work.Description),
				Thumbnail:   discordImage{URL: toyboxLogoURL},
				Color:       discordEmbedColor,
				Image:       discordImage{URL: work.ThumbnailURL},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(body))
	if err != nil {
		return errors.New("failed to create Discord webhook request")
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := n.client.Do(req)
	if err != nil {
		return errors.New("failed to request Discord webhook")
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Discord webhook returned status %d", res.StatusCode)
	}

	return nil
}

func truncateDescription(description string) string {
	if utf8.RuneCountInString(description) <= maxDescriptionRune {
		return description
	}

	runes := []rune(description)
	return string(runes[:maxDescriptionRune]) + "…"
}
